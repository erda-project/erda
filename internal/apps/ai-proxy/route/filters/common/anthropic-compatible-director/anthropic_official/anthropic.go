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

package anthropic_official

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/anthropic-compatible-director/common/message_converter"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/anthropic-compatible-director/common/openai_extended"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

const (
	APIVendor api_style.APIVendor = "anthropic"
)

type AnthropicRequest struct {
	message_converter.BaseAnthropicRequest
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

type AnthropicDirector struct {
	StreamMessageInfo message_converter.AnthropicStreamMessageInfo
}

func NewDirector() *AnthropicDirector {
	return &AnthropicDirector{}
}

func (f *AnthropicDirector) OfficialDirector(pr *httputil.ProxyRequest, apiStyleConfig api_style.APIStyleConfig) error {
	model := ctxhelper.MustGetModel(pr.Out.Context())
	// openai format request
	var openaiReq openai_extended.OpenAIRequestExtended
	if err := json.NewDecoder(pr.Out.Body).Decode(&openaiReq); err != nil {
		panic(fmt.Errorf("failed to decode request body as openai format, err: %v", err))
	}
	// convert to: anthropic format request
	baseAnthropicReq := message_converter.ConvertOpenAIRequestToBaseAnthropicRequest(openaiReq)
	anthropicReq := AnthropicRequest{
		BaseAnthropicRequest: baseAnthropicReq,
		Model:                model.Metadata.Public["model_name"].GetStringValue(),
		Stream:               openaiReq.Stream,
	}

	anthropicReqBytes, err := json.Marshal(&anthropicReq)
	if err != nil {
		panic(fmt.Errorf("failed to marshal anthropic request: %v", err))
	}
	if err := body_util.SetBody(pr.Out, anthropicReqBytes); err != nil {
		return fmt.Errorf("failed to set anthropic request body: %v", err)
	}

	return nil
}

func (f *AnthropicDirector) OnBodyChunk(resp *http.Response, chunk []byte, index int64) ([]byte, error) {
	// non-stream
	if !ctxhelper.MustGetIsStream(resp.Request.Context()) {
		// convert all at once
		var anthropicResp message_converter.AnthropicResponse
		if err := json.Unmarshal(chunk, &anthropicResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %s, err: %v", string(chunk), err)
		}
		openaiResp, err := anthropicResp.ConvertToOpenAIFormat(ctxhelper.MustGetModel(resp.Request.Context()).Metadata.Public["model_id"].GetStringValue())
		if err != nil {
			return nil, fmt.Errorf("failed to convert anthropic response body to openai format, err: %v", err)
		}
		return json.Marshal(openaiResp)
	}
	// stream
	var chunkWriter bytes.Buffer
	openaiChunks, err := f.pipeAnthropicStream(resp.Request.Context(), io.NopCloser(bytes.NewBuffer(chunk)), &chunkWriter)
	if err != nil {
		return nil, fmt.Errorf("failed to parse anthropic eventstream, err: %v", err)
	}
	var chunkDataList [][]byte
	for _, openaiChunk := range openaiChunks {
		b, err := json.Marshal(openaiChunk)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal openai chunk, err: %v", err)
		}
		chunkData := vars.ConcatChunkDataPrefix(b)
		chunkDataList = append(chunkDataList, chunkData)
	}
	result := bytes.Join(chunkDataList, []byte("\n\n"))
	return result, nil
}

func (f *AnthropicDirector) OnComplete(resp *http.Response) ([]byte, error) {
	if ctxhelper.MustGetIsStream(resp.Request.Context()) {
		// append [DONE] chunk
		doneChunk := vars.ConcatChunkDataPrefix([]byte("[DONE]"))
		return doneChunk, nil
	}
	return nil, nil
}
