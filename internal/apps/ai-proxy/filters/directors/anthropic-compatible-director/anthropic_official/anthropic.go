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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/anthropic-compatible-director/common/message_converter"
	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/anthropic-compatible-director/common/openai_extended"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
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
	*reverseproxy.DefaultResponseFilter

	StreamMessageInfo message_converter.AnthropicStreamMessageInfo
}

func NewDirector() *AnthropicDirector {
	return &AnthropicDirector{DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter()}
}

func (f *AnthropicDirector) OfficialDirector(ctx context.Context, infor reverseproxy.HttpInfor, apiStyleConfig api_style.APIStyleConfig) error {
	model := ctxhelper.MustGetModel(ctx)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		// openai format request
		var openaiReq openai_extended.OpenAIRequestExtended
		if err := json.NewDecoder(infor.Body()).Decode(&openaiReq); err != nil {
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
		infor.SetBody2(anthropicReqBytes)
	})

	return nil
}

func (f *AnthropicDirector) OnResponseChunk(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (signal reverseproxy.Signal, err error) {
	// non-stream
	if !ctxhelper.GetIsStream(ctx) {
		f.DefaultResponseFilter.Buffer.Write(chunk) // write to buffer, so we can get allChunks later
		return reverseproxy.Continue, nil
	}
	// stream
	var chunkWriter bytes.Buffer
	openaiChunks, err := f.pipeAnthropicStream(ctx, io.NopCloser(bytes.NewBuffer(chunk)), &chunkWriter)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to parse anthropic eventstream, err: %v", err)
	}
	for _, openaiChunk := range openaiChunks {
		b, err := json.Marshal(openaiChunk)
		if err != nil {
			return reverseproxy.Intercept, fmt.Errorf("failed to marshal openai chunk, err: %v", err)
		}
		chunkData := vars.ConcatChunkDataPrefix(b)
		if _, err := w.Write(chunkData); err != nil {
			return reverseproxy.Intercept, fmt.Errorf("failed to write openai chunk, err: %v", err)
		}
	}
	return reverseproxy.Continue, nil
}

func (f *AnthropicDirector) OnResponseEOF(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (err error) {
	// only stream style need append [DONE] chunk
	if !ctxhelper.GetIsStream(ctx) {
		// convert all at once
		var anthropicResp message_converter.AnthropicResponse
		if err := json.Unmarshal(f.DefaultResponseFilter.Buffer.Bytes(), &anthropicResp); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %s, err: %v", string(chunk), err)
		}
		openaiResp, err := anthropicResp.ConvertToOpenAIFormat(ctxhelper.MustGetModel(ctx).Metadata.Public["model_id"].GetStringValue())
		if err != nil {
			return fmt.Errorf("failed to convert anthropic response body to openai format, err: %v", err)
		}
		b, err := json.Marshal(openaiResp)
		if err != nil {
			return fmt.Errorf("failed to marshal openai resp, err: %v", err)
		}
		infor.Header().Del("Content-Length")
		return f.DefaultResponseFilter.OnResponseEOF(ctx, nil, w, b)
	}
	// append [DONE] chunk
	doneChunk := vars.ConcatChunkDataPrefix([]byte("[DONE]"))
	if _, err := w.Write(doneChunk); err != nil {
		return fmt.Errorf("failed to write openai chunk, err: %v", err)
	}
	return nil
}
