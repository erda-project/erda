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
	EncodingFormat string                           `json:"encoding_format,omitempty"`
	Output         *CanonicalMultimodalOutputConfig `json:"output,omitempty"`
}

type CanonicalMultimodalInputItem struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	VideoURL string `json:"video_url,omitempty"`
}

type CanonicalMultimodalOutputConfig struct {
	Dense  *bool `json:"dense,omitempty"`
	Multi  *bool `json:"multi,omitempty"`
	Sparse *bool `json:"sparse,omitempty"`
	Fusion *bool `json:"fusion,omitempty"`
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
		EncodingFormat: req.EncodingFormat,
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

	if req.Output != nil {
		if isTrue(req.Output.Multi) {
			arkReq.MultiEmbedding = &VolcengineEmbeddingSwitch{Type: "enabled"}
		}
		if isTrue(req.Output.Sparse) {
			arkReq.SparseEmbedding = &VolcengineEmbeddingSwitch{Type: "enabled"}
		}
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

func isTrue(v *bool) bool {
	return v != nil && *v
}

func isSupportedDimension(v int) bool {
	return v == dimension1024 || v == defaultDimensions
}
