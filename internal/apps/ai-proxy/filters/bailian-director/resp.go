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

package bailian_director

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
)

func (f *BailianDirector) OnResponseChunk(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (signal reverseproxy.Signal, err error) {
	isStream, _ := ctxhelper.GetIsStream(ctx)
	if !isStream {
		f.DefaultResponseFilter.Buffer.Write(chunk)
		return reverseproxy.Continue, nil
	}
	// handle chunk
	return f.handleResponseStreamChunk(ctx, w, chunk)
}

// chunk format:
// ```plain text
// data: {"Data":{"Debug":{},"DocReferences":[],"ResponseId":"ba91dde7238a401383dff1fda7c950e6","Text":"hello","Thoughts":[]},"RequestId":"db1190f8ff344bdca363f05420fa6b58","Success":true}
//
// data: {"Data":{"Debug":{},"DocReferences":[],"ResponseId":"ba91dde7238a401383dff1fda7c950e6","Text":"hello world","Thoughts":[]},"RequestId":"db1190f8ff344bdca363f05420fa6b58","Success":true}
//
// data: {"Data":{"Debug":{},"DocReferences":[],"ResponseId":"ba91dde7238
// ```
// the last line may be incomplete.
func (f *BailianDirector) handleResponseStreamChunk(ctx context.Context, w reverseproxy.Writer, bailianChunk []byte) (reverseproxy.Signal, error) {
	f.DefaultResponseFilter.Buffer.Write(bailianChunk)
	signal, lastCompleteDeltaResp := f.getDeltaTextFromStreamChunk(f.DefaultResponseFilter.Buffer.Bytes())
	if signal == reverseproxy.Intercept || lastCompleteDeltaResp == nil {
		return signal, nil
	}
	// convert bailian response to openai response
	model, _ := ctxhelper.GetModel(ctx)
	openaiResp, err := convertToOpenaiStreamChunk(model, *lastCompleteDeltaResp)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to convert bailian response to openai response, err: %v", err)
	}
	b, err := json.Marshal(openaiResp)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to marshal openai response, err: %v", err)
	}
	b = vars.ConcatChunkDataPrefix(b)
	if _, err := w.Write(b); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to write openai response, err: %v", err)
	}
	return reverseproxy.Continue, nil
}

func (f *BailianDirector) getDeltaTextFromStreamChunk(allChunks []byte) (reverseproxy.Signal, *CompletionResponse) {
	allLines := string(allChunks)
	allDataLines := strutil.Split(allLines, "\n", true)
	// try to get the last complete data line
	var lastCompleteDataLine CompletionResponse
	var lastCompleteDataLineText string
	for i := len(allDataLines) - 1; i >= 0; i-- {
		if i < f.lastCompletionDataLineIndex {
			break
		}
		if err := json.Unmarshal(vars.TrimChunkDataPrefix([]byte(allDataLines[i])), &lastCompleteDataLine); err == nil {
			f.lastCompletionDataLineIndex = i
			if lastCompleteDataLine.Success == false {
				return reverseproxy.Intercept, nil
			}
			if lastCompleteDataLine.Data == nil {
				continue
			}
			if lastCompleteDataLine.Data.Text == nil || *lastCompleteDataLine.Data.Text == "" {
				continue
			}
			lastCompleteDataLineText = *lastCompleteDataLine.Data.Text
			break
		}
	}
	if lastCompleteDataLineText == "" {
		return reverseproxy.Continue, nil
	}
	// 因为 qwenv1 目前 stream 不是真 chunk，每个 data line 的 text 都是在之前的基础上进行 append，所以需要主动计算增量值，模拟真 chunk 形式返回
	delta := strings.TrimPrefix(lastCompleteDataLineText, f.lastCompletionDataLineText)
	if delta == "" {
		return reverseproxy.Continue, nil
	}
	f.lastCompletionDataLineText = lastCompleteDataLineText
	lastCompleteDataLine.Data.Text = &delta
	return reverseproxy.Continue, &lastCompleteDataLine
}

func (f *BailianDirector) OnResponseEOF(ctx context.Context, _ reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (err error) {
	// only stream style need append [DONE] chunk
	isStream, _ := ctxhelper.GetIsStream(ctx)
	if !isStream {
		// convert all at once
		model, _ := ctxhelper.GetModel(ctx)
		var bailianResp CompletionResponse
		if err := json.Unmarshal(f.DefaultResponseFilter.Buffer.Bytes(), &bailianResp); err != nil {
			return fmt.Errorf("failed to unmarshal bailian response, err: %v", err)
		}
		openaiResp, err := convertToOpenaiResponse(model, bailianResp)
		if err != nil {
			return fmt.Errorf("failed to convert bailian response to openai response, err: %v", err)
		}
		b, err := json.Marshal(openaiResp)
		if err != nil {
			return fmt.Errorf("failed to marshal openai response, err: %v", err)
		}
		return f.DefaultResponseFilter.OnResponseEOF(ctx, nil, w, b)
	}
	// append [DONE] chunk
	doneChunk := vars.ConcatChunkDataPrefix([]byte("[DONE]"))
	if _, err := w.Write(doneChunk); err != nil {
		return fmt.Errorf("failed to write openai response, err: %v", err)
	}
	return nil
}

func convertToOpenaiStreamChunk(model *modelpb.Model, bailianResponse CompletionResponse) (*openai.ChatCompletionStreamResponse, error) {
	openaiResp := &openai.ChatCompletionStreamResponse{
		ID:     *bailianResponse.Data.ResponseId,
		Object: "chat.completion.chunk",
		Model:  model.Name,
		Choices: []openai.ChatCompletionStreamChoice{
			{
				Index: 0,
				Delta: openai.ChatCompletionStreamChoiceDelta{
					Role:    openai.ChatMessageRoleAssistant,
					Content: *bailianResponse.Data.Text,
				},
				FinishReason: openai.FinishReasonNull,
			},
		},
	}
	return openaiResp, nil
}

func convertToOpenaiResponse(model *modelpb.Model, bailianResponse CompletionResponse) (*openai.CompletionResponse, error) {
	openaiResp := &openai.CompletionResponse{
		ID:     *bailianResponse.Data.ResponseId,
		Object: "chat.completion",
		Model:  model.Name,
		Choices: []openai.CompletionChoice{
			{
				Text:         *bailianResponse.Data.Text,
				Index:        0,
				FinishReason: string(openai.FinishReasonStop),
			},
		},
	}
	return openaiResp, nil
}
