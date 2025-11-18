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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"google.golang.org/genai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_segment_getter"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

var (
	_ filter_define.ProxyResponseModifier = (*GoogleVertexAIDirectorResponse)(nil)
)

var ResponseModifierCreator filter_define.ResponseModifierCreator = func(name string, _ json.RawMessage) filter_define.ProxyResponseModifier {
	return &GoogleVertexAIDirectorResponse{}
}

func init() {
	filter_define.RegisterFilterCreator("google-vertex-ai-director", ResponseModifierCreator)
}

type GoogleVertexAIDirectorResponse struct {
	filter_define.PassThroughResponseModifier
}

func (f *GoogleVertexAIDirectorResponse) Enable(resp *http.Response) bool {
	apiSegment := api_segment_getter.GetAPISegment(resp.Request.Context())
	return apiSegment != nil && strings.EqualFold(string(apiSegment.APIStyle), string(api_style.APIStyleGoogleVertexAI))
}

func (f *GoogleVertexAIDirectorResponse) OnBodyChunk(resp *http.Response, chunk []byte, index int64) ([]byte, error) {
	// distinguish by if it's openai-compatible or not
	pathMatcher := ctxhelper.MustGetPathMatcher(resp.Request.Context())

	// openai-compatible
	if pathMatcher.Match(vars.RequestPathPrefixV1ChatCompletions) {
		return chunk, nil
	}

	// not openai-compatible, only support non-stream response yet
	if ctxhelper.MustGetIsStream(resp.Request.Context()) {
		return nil, fmt.Errorf("streaming response is not support yet")
	}

	// convert google vertex ai response to openai style
	if pathMatcher.Match(vars.RequestPathPrefixV1ImagesGenerations) || pathMatcher.Match(vars.RequestPathPrefixV1ImagesEdits) {
		return convertGenAIResponseToOpenAIImageGenerationResponse(chunk)
	}
	return nil, fmt.Errorf("unsupported path: %s", pathMatcher.Pattern)
}

func (f *GoogleVertexAIDirectorResponse) OnComplete(resp *http.Response) ([]byte, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	// stream
	if ctxhelper.MustGetIsStream(resp.Request.Context()) {
		// append [DONE] chunk
		doneChunk := vars.ConcatChunkDataPrefix([]byte("[DONE]"))
		return doneChunk, nil
	}
	return nil, nil
}

func convertGenAIResponseToOpenAIImageGenerationResponse(resp []byte) ([]byte, error) {
	var gr genai.GenerateContentResponse
	if err := json.Unmarshal(resp, &gr); err != nil {
		return nil, fmt.Errorf("failed to decode vertex generateContent response: %w", err)
	}

	out := openai.ImageResponse{Created: time.Now().Unix()}

	for _, c := range gr.Candidates {
		for _, p := range c.Content.Parts {
			if p.InlineData != nil && len(p.InlineData.Data) > 0 {
				encoded := base64.StdEncoding.EncodeToString(p.InlineData.Data)
				out.Data = append(out.Data, openai.ImageResponseDataInner{B64JSON: encoded})
				continue
			}
			if p.FileData != nil && p.FileData.FileURI != "" {
				out.Data = append(out.Data, openai.ImageResponseDataInner{URL: p.FileData.FileURI})
			}
		}
	}

	if len(out.Data) == 0 {
		return nil, fmt.Errorf("no image parts found in vertex response")
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("failed to encode openai images response: %w", err)
	}
	return b, nil
}
