// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package volcengine_ark

import (
	"encoding/json"
	"fmt"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
)

const (
	arkMultimodalEmbeddingPath = "/api/v3/embeddings/multimodal"
	dimension1024              = 1024
	defaultDimensions          = 2048
	defaultEncodingFormat      = "float"
	defaultInstructions        = "Target_modality: text and video.\nInstruction:Compress the text/video into one word.\nQuery:"
)

type CanonicalMultimodalEmbeddingRequest struct {
	Model          string                           `json:"model,omitempty"`
	Input          []CanonicalMultimodalInputItem   `json:"input"`
	Dimensions     *int                             `json:"dimensions,omitempty"`
	Instruction    string                           `json:"instruction,omitempty"`
	EncodingFormat string                           `json:"encoding_format,omitempty"` // legacy compatibility
	Output         *CanonicalMultimodalOutputConfig `json:"output,omitempty"`
	Options        map[string]any                   `json:"options,omitempty"`
}

type CanonicalMultimodalInputItem struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	VideoURL string `json:"video_url,omitempty"`
}

type CanonicalMultimodalOutputConfig struct {
	Primary    string   `json:"primary,omitempty"`
	Additional []string `json:"additional,omitempty"`
}

type VolcengineMultimodalEmbeddingRequest struct {
	Model           string                     `json:"model,omitempty"`
	Instructions    string                     `json:"instructions,omitempty"`
	Input           []map[string]any           `json:"input"`
	Dimensions      int                        `json:"dimensions"`
	MultiEmbedding  *VolcengineEmbeddingSwitch `json:"multi_embedding,omitempty"`
	SparseEmbedding *VolcengineEmbeddingSwitch `json:"sparse_embedding,omitempty"`
	EncodingFormat  string                     `json:"encoding_format,omitempty"`
}

type VolcengineEmbeddingSwitch struct {
	Type string `json:"type"`
}

func (f *VolcengineMultimodalEmbeddingConverter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	var req CanonicalMultimodalEmbeddingRequest
	if err := json.NewDecoder(pr.Out.Body).Decode(&req); err != nil {
		return fmt.Errorf("failed to decode multimodal embedding request: %w", err)
	}
	if len(req.Input) == 0 {
		return fmt.Errorf("input is required")
	}

	items := make([]map[string]any, 0, len(req.Input))
	for i, item := range req.Input {
		converted, err := convertInputItem(item)
		if err != nil {
			return fmt.Errorf("input[%d]: %w", i, err)
		}
		items = append(items, converted)
	}

	arkReq := &VolcengineMultimodalEmbeddingRequest{
		Model:          req.Model,
		Instructions:   strings.TrimSpace(req.Instruction),
		Input:          items,
		Dimensions:     defaultDimensions,
		EncodingFormat: resolveEncodingFormat(req),
	}
	if req.Dimensions != nil {
		if !isSupportedDimension(*req.Dimensions) {
			return fmt.Errorf("dimensions must be one of [1024, 2048]")
		}
		arkReq.Dimensions = *req.Dimensions
	}
	if strings.TrimSpace(arkReq.Instructions) == "" {
		arkReq.Instructions = defaultInstructions
	}
	if strings.TrimSpace(arkReq.EncodingFormat) == "" {
		arkReq.EncodingFormat = defaultEncodingFormat
	}

	if err := applyOutputConfig(req.Output, req.Input, arkReq); err != nil {
		return err
	}

	pr.Out.URL.Path = arkMultimodalEmbeddingPath
	return body_util.SetBody(pr.Out, arkReq)
}

func convertInputItem(item CanonicalMultimodalInputItem) (map[string]any, error) {
	t := strings.ToLower(strings.TrimSpace(item.Type))
	switch t {
	case "text":
		if strings.TrimSpace(item.Text) == "" {
			return nil, fmt.Errorf("text is required when type=text")
		}
		return map[string]any{"type": "text", "text": item.Text}, nil
	case "image":
		if strings.TrimSpace(item.ImageURL) == "" {
			return nil, fmt.Errorf("image_url is required when type=image")
		}
		return map[string]any{"type": "image", "image_url": item.ImageURL}, nil
	case "video":
		if strings.TrimSpace(item.VideoURL) == "" {
			return nil, fmt.Errorf("video_url is required when type=video")
		}
		return map[string]any{"type": "video", "video_url": item.VideoURL}, nil
	default:
		return nil, fmt.Errorf("unsupported type %q", item.Type)
	}
}

func resolveEncodingFormat(req CanonicalMultimodalEmbeddingRequest) string {
	legacy := strings.TrimSpace(req.EncodingFormat)
	if req.Options == nil {
		return legacy
	}
	if v, ok := req.Options["encoding_format"]; ok {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return legacy
}

func applyOutputConfig(output *CanonicalMultimodalOutputConfig, input []CanonicalMultimodalInputItem, arkReq *VolcengineMultimodalEmbeddingRequest) error {
	// Optional output defaults to dense only, and dense is always present in Ark response.
	if output == nil {
		return nil
	}

	primary := strings.ToLower(strings.TrimSpace(output.Primary))
	if primary == "" {
		primary = "dense"
	}
	if primary != "dense" && primary != "fusion" {
		return fmt.Errorf("output.primary must be one of [dense, fusion]")
	}
	if primary == "fusion" {
		if len(output.Additional) > 0 {
			return fmt.Errorf("output.additional must be empty when output.primary=fusion")
		}
		return fmt.Errorf("output.primary=fusion is not supported by volcengine-ark multimodal embedding")
	}

	additional := map[string]bool{}
	for _, v := range output.Additional {
		key := strings.ToLower(strings.TrimSpace(v))
		if key == "" {
			continue
		}
		if key != "multi" && key != "sparse" {
			return fmt.Errorf("output.additional must contain only [multi, sparse]")
		}
		additional[key] = true
	}

	if additional["multi"] {
		arkReq.MultiEmbedding = &VolcengineEmbeddingSwitch{Type: "enabled"}
	}
	if additional["sparse"] {
		for _, item := range input {
			if strings.ToLower(strings.TrimSpace(item.Type)) != "text" {
				return fmt.Errorf("output.additional=sparse is only supported for text input")
			}
		}
		arkReq.SparseEmbedding = &VolcengineEmbeddingSwitch{Type: "enabled"}
	}
	return nil
}

func isSupportedDimension(v int) bool {
	return v == dimension1024 || v == defaultDimensions
}
