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

package google_vertex_ai_director

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime"
	"mime/multipart"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_segment_getter"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	custom_http_director "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/custom-http-director"
)

const (
	placeholderKeyGCPAccessToken = "gcp-sa-access-token"
	placeholderKeyGCPProjectID   = "gcp-project-id"
)

type GoogleVertexAIDirector struct {
	*custom_http_director.CustomHTTPDirector
}

var (
	_ filter_define.ProxyRequestRewriter = (*GoogleVertexAIDirector)(nil)
)

var Creator filter_define.RequestRewriterCreator = func(name string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &GoogleVertexAIDirector{CustomHTTPDirector: custom_http_director.New()}
}

func init() {
	filter_define.RegisterFilterCreator("google-vertex-ai-director", Creator)
}

func (f *GoogleVertexAIDirector) Enable(pr *httputil.ProxyRequest) bool {
	apiSegment := api_segment_getter.GetAPISegment(pr.In.Context())
	return apiSegment != nil &&
		strings.EqualFold(string(apiSegment.APIStyle), string(api_style.APIStyleGoogleVertexAI))
}

func (f *GoogleVertexAIDirector) OnProxyRequest(pr *httputil.ProxyRequest) error {
	if !f.Enable(pr) {
		return nil
	}
	// set custom placeholders for service-provider to render apiStyleConfig
	sp := ctxhelper.MustGetServiceProvider(pr.In.Context())
	ak, err := getGCPAccessToken(pr.In.Context(), sp)
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}
	sp.Metadata.Secret[placeholderKeyGCPAccessToken] = structpb.NewStringValue(ak)
	gcpProjectID, err := getGCPProjectId(pr.In.Context(), sp)
	if err != nil {
		return fmt.Errorf("failed to get gcp project id: %v", err)
	}
	sp.Metadata.Public[placeholderKeyGCPProjectID] = structpb.NewStringValue(gcpProjectID)
	// do unified custom http director
	if err := f.CustomHTTPDirector.OnProxyRequest(pr); err != nil {
		return fmt.Errorf("vertex-ai custom-http-director error: %w", err)
	}

	// do custom mapping
	pathMatcher := ctxhelper.MustGetPathMatcher(pr.In.Context())
	if pathMatcher.Match(common.RequestPathPrefixV1ImagesGenerations) {
		if err := mappingImageGenerationAPI(pr); err != nil {
			return err
		}
	} else if pathMatcher.Match(common.RequestPathPrefixV1ImagesEdits) {
		if err := mappingImageEditAPI(pr); err != nil {
			return err
		}
	}

	return nil
}

// see: https://docs.cloud.google.com/vertex-ai/generative-ai/docs/model-reference/inference?hl=zh-cn
// see: https://docs.cloud.google.com/vertex-ai/generative-ai/docs/multimodal/image-generation?hl=zh-cn#googlegenaisdk_imggen_mmflash_with_txt-drest
func mappingImageGenerationAPI(pr *httputil.ProxyRequest) error {
	// mapping request body from openai style to google vertex ai style
	var req openai.ImageRequest
	if err := json.NewDecoder(pr.Out.Body).Decode(&req); err != nil {
		return fmt.Errorf("failed to decode openai image request: %w", err)
	}

	vertexReq := struct {
		Contents              []*genai.Content            `json:"contents"`
		GenerateContentConfig genai.GenerateContentConfig `json:"generationConfig"`
	}{
		Contents: []*genai.Content{
			{
				Role: genai.RoleUser,
				Parts: []*genai.Part{
					{Text: req.Prompt},
				},
			},
		},
		GenerateContentConfig: genai.GenerateContentConfig{
			ResponseModalities: []string{string(genai.ModalityImage)},
			ImageConfig: &genai.ImageConfig{
				ImageSize:   convertImageSize(req),
				AspectRatio: convertAspectRatio(req),
			},
		},
	}

	if err := body_util.SetBody(pr.Out, vertexReq); err != nil {
		return fmt.Errorf("failed to set transformed vertex body: %w", err)
	}
	return nil
}

func mappingImageEditAPI(pr *httputil.ProxyRequest) error {
	payload, err := parseImageEditMultipart(pr.Out.Body, pr.Out.Header.Get("Content-Type"))
	if err != nil {
		return fmt.Errorf("failed to parse multipart image edit request: %w", err)
	}
	if len(payload.Image) == 0 {
		return fmt.Errorf("missing image part in multipart image edit request")
	}
	imageMime := payload.ImageContentType
	if imageMime == "" {
		imageMime = "image/png"
	}

	editReq := openai.ImageRequest{
		Prompt:  payload.Prompt,
		Size:    payload.Size,
		Quality: payload.Quality,
	}
	vertexReq := struct {
		Contents              []*genai.Content            `json:"contents"`
		GenerateContentConfig genai.GenerateContentConfig `json:"generationConfig"`
	}{
		Contents: []*genai.Content{
			{
				Role: genai.RoleUser,
				Parts: []*genai.Part{
					{InlineData: &genai.Blob{
						Data:        payload.Image,
						DisplayName: "",
						MIMEType:    imageMime,
					}},
					{Text: payload.Prompt},
				},
			},
		},
		GenerateContentConfig: genai.GenerateContentConfig{
			ResponseModalities: []string{string(genai.ModalityImage)},
			ImageConfig: &genai.ImageConfig{
				ImageSize:   convertImageSize(editReq),
				AspectRatio: convertAspectRatio(editReq),
			},
		},
	}

	if err := body_util.SetBody(pr.Out, vertexReq); err != nil {
		return fmt.Errorf("failed to set transformed vertex body: %w", err)
	}

	// the original content-type is multipart/form-data
	pr.Out.Header.Set("Content-Type", "application/json")

	return nil
}

type imageEditMultipartPayload struct {
	Prompt           string
	Image            []byte
	ImageContentType string
	Size             string
	Quality          string
}

func parseImageEditMultipart(body io.Reader, contentType string) (*imageEditMultipartPayload, error) {
	if body == nil {
		return nil, fmt.Errorf("request body is empty")
	}
	if contentType == "" {
		return nil, fmt.Errorf("content-type header is empty")
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content-type: %w", err)
	}
	if !strings.EqualFold(mediaType, "multipart/form-data") {
		return nil, fmt.Errorf("unsupported content-type %q", mediaType)
	}
	boundary, ok := params["boundary"]
	if !ok {
		return nil, fmt.Errorf("multipart boundary not found")
	}

	reader := multipart.NewReader(body, boundary)
	payload := &imageEditMultipartPayload{}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate multipart data: %w", err)
		}

		if err := func(part *multipart.Part) error {
			defer func() { _ = part.Close() }()
			name := part.FormName()
			switch {
			case name == "prompt":
				data, err := io.ReadAll(part)
				if err != nil {
					return fmt.Errorf("failed to read prompt part: %w", err)
				}
				payload.Prompt = string(data)
			case name == "size":
				data, err := io.ReadAll(part)
				if err != nil {
					return fmt.Errorf("failed to read size part: %w", err)
				}
				payload.Size = string(data)
			case name == "quality":
				data, err := io.ReadAll(part)
				if err != nil {
					return fmt.Errorf("failed to read quality part: %w", err)
				}
				payload.Quality = string(data)
			case isImageFormField(name):
				if len(payload.Image) == 0 {
					ct := part.Header.Get("Content-Type")
					data, err := io.ReadAll(part)
					if err != nil {
						return fmt.Errorf("failed to read image part: %w", err)
					}
					payload.Image = data
					payload.ImageContentType = ct
				} else {
					if _, err := io.Copy(io.Discard, part); err != nil {
						return fmt.Errorf("failed to discard duplicate image part: %w", err)
					}
				}
			default:
				if _, err := io.Copy(io.Discard, part); err != nil {
					return fmt.Errorf("failed to discard part %q: %w", name, err)
				}
			}
			return nil
		}(part); err != nil {
			return nil, err
		}
	}

	return payload, nil
}

func isImageFormField(name string) bool {
	if name == "" {
		return false
	}
	if name == "image" || name == "image[]" {
		return true
	}
	return strings.HasPrefix(name, "image[")
}

func convertImageSize(req openai.ImageRequest) string {
	quality := strings.ToLower(req.Quality)
	maxDim := maxDimensionFromSize(req.Size)

	switch quality {
	case openai.CreateImageQualityHD, openai.CreateImageQualityHigh:
		if maxDim >= 2048 {
			return "4K"
		}
		return "2K"
	case openai.CreateImageQualityMedium:
		if maxDim >= 1536 {
			return "2K"
		}
		return "1K"
	case openai.CreateImageQualityLow:
		return "1K"
	}

	if maxDim >= 2048 {
		return "4K"
	}
	if maxDim >= 1536 {
		return "2K"
	}
	return "1K"
}

func convertAspectRatio(req openai.ImageRequest) string {
	ratio := ratioFromSize(req.Size)
	targets := []struct {
		label string
		value float64
	}{
		{"1:1", 1},
		{"2:3", 2.0 / 3.0},
		{"3:2", 3.0 / 2.0},
		{"3:4", 3.0 / 4.0},
		{"4:3", 4.0 / 3.0},
		{"9:16", 9.0 / 16.0},
		{"16:9", 16.0 / 9.0},
		{"21:9", 21.0 / 9.0},
	}

	best := "1:1"
	bestDiff := math.MaxFloat64
	for _, target := range targets {
		diff := math.Abs(ratio - target.value)
		if diff < bestDiff {
			bestDiff = diff
			best = target.label
		}
	}
	return best
}

func ratioFromSize(size string) float64 {
	width, height := parseImageDimensions(size)
	if width == 0 || height == 0 {
		return 1
	}
	return float64(width) / float64(height)
}

func maxDimensionFromSize(size string) int {
	width, height := parseImageDimensions(size)
	if width == 0 && height == 0 {
		return 0
	}
	if width > height {
		return width
	}
	return height
}

func parseImageDimensions(size string) (int, int) {
	if size == "" {
		return 0, 0
	}
	parts := strings.Split(size, "x")
	if len(parts) != 2 {
		return 0, 0
	}
	width, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0
	}
	height, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0
	}
	return width, height
}
