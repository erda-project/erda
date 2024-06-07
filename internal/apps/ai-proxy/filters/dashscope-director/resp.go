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

package dashscope_director

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func (f *DashScopeDirector) OnResponseChunk(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (signal reverseproxy.Signal, err error) {
	model := ctxhelper.MustGetModel(ctx)
	modelMeta := getModelMeta(model)

	// switch by model name
	switch modelMeta.Public.ModelName {
	case metadata.AliyunDashScopeModelNameQwenLong:
		return f.DefaultResponseFilter.OnResponseChunk(ctx, infor, w, chunk)
	case metadata.AliyunDashScopeModelNameQwenVLPlus, metadata.AliyunDashScopeModelNameQwenVLMax:
		return f.qwenVLOnResponseChunk(ctx, infor, w, chunk)
	default:
		return reverseproxy.Intercept, fmt.Errorf("unsupported model: %s", model.Name)
	}
}

func (f *DashScopeDirector) qwenVLOnResponseChunk(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (signal reverseproxy.Signal, err error) {
	if !ctxhelper.GetIsStream(ctx) {
		f.DefaultResponseFilter.Buffer.Write(chunk)
		return reverseproxy.Continue, nil
	}
	// handle chunk
	return f.qwenVLHandleResponseStreamChunk(ctx, w, chunk)
}

type (
	QwenVLRespStreamChunk struct {
		Output    QwenVLRespStreamChunkOutput `json:"output"`
		RequestID string                      `json:"request_id"`
	}
	QwenVLRespStreamChunkOutput struct {
		Choices []QwenVLRespStreamChunkOutputChoice `json:"choices"`
	}
	QwenVLRespStreamChunkOutputChoice struct {
		Message      QwenVLRespStreamChunkOutputChoiceMessage `json:"message"`
		FinishReason string                                   `json:"finish_reason"`
	}
	QwenVLRespStreamChunkOutputChoiceMessage struct {
		Content []QwenVLRespStreamChunkOutputChoiceMessageContent `json:"content"`
		Role    string                                            `json:"role"`
	}
	QwenVLRespStreamChunkOutputChoiceMessageContent struct {
		Text string `json:"text"`
	}
)

// vl resp stream format
// id:1
// event:result
// :HTTP_STATUS/200
// data:{"output":{"choices":[{"message":{"content":[{"text":"Hello"}],"role":"assistant"},"finish_reason":"null"}]},"usage":{"input_tokens":92,"output_tokens":1},"request_id":"e2edfd5d-f5c3-958b-b058-f8341b31af05"}
//
// id:5
// event:result
// :HTTP_STATUS/200
// data:{"output":{"choices":[{"message":{"content":[{"text":"Hello! How can I assist you today?"}],"role":"assistant"},"finish_reason":"stop"}]},"usage":{"input_tokens":92,"output_tokens":10},"request_id":"e2edfd5d-f5c3-958b-b058-f8341b31af05"}
func (f *DashScopeDirector) qwenVLHandleResponseStreamChunk(ctx context.Context, w reverseproxy.Writer, qwenVLChunk []byte) (reverseproxy.Signal, error) {
	model := ctxhelper.MustGetModel(ctx)
	modelMeta := getModelMeta(model)
	modelName := modelMeta.Public.ModelName

	// 因为 qwen 目前 stream 不是真 chunk，每个 data line 的 text 都是在之前的基础上进行 append，所以需要主动计算增量值，模拟真 chunk 形式返回
	// 所以，只需要找到 chunk 里最后一个 data line，然后计算增量值，转换为一个 OpenAI Chunk 即可
	signal, lastCompleteDeltaResp := f.getDeltaTextFromStreamChunk(qwenVLChunk)
	if signal == reverseproxy.Intercept || lastCompleteDeltaResp == nil {
		return signal, nil
	}
	// convert qwenVL response to openai response
	openaiChunk, err := convertToOpenAIStreamChunk(*lastCompleteDeltaResp, string(modelName))
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to convert qwenVL response to openai response, err: %v", err)
	}
	b, err := json.Marshal(&openaiChunk)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to marshal openai response, err: %v", err)
	}
	b = vars.ConcatChunkDataPrefix(b)
	if _, err := w.Write(b); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to write openai response, err: %v", err)
	}
	return reverseproxy.Continue, nil
}

// vl resp non-stream format
// {"output":{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":[{"text":"Hello! How can I assist you today?"}]}}]},"usage":{"output_tokens":10,"input_tokens":92},"request_id":"cc7bcb16-0a9d-9ca4-ae78-b792393eb840"}

func (f *DashScopeDirector) getDeltaTextFromStreamChunk(allChunks []byte) (reverseproxy.Signal, *QwenVLRespStreamChunk) {
	// 按行解析，保留 `data:` 开头的行
	var allDataLines []string
	scanner := bufio.NewScanner(bytes.NewReader(allChunks))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		//// remove prefix `data:`
		//line = strings.TrimPrefix(line, "data:")
		allDataLines = append(allDataLines, line)
	}
	// try to get the last complete data line
	var lastCompleteDataLine QwenVLRespStreamChunk
	var lastChoice QwenVLRespStreamChunkOutputChoice
	var lastCompleteDataLineText string
	var lastContentItem QwenVLRespStreamChunkOutputChoiceMessageContent
	for i := len(allDataLines) - 1; i >= 0; i-- {
		if i < f.lastCompletionDataLineIndex {
			break
		}
		if err := json.Unmarshal(TrimChunkDataPrefix([]byte(allDataLines[i])), &lastCompleteDataLine); err == nil {
			f.lastCompletionDataLineIndex = i
			if len(lastCompleteDataLine.Output.Choices) == 0 {
				continue
			}
			// get last choice
			lastChoice = lastCompleteDataLine.Output.Choices[len(lastCompleteDataLine.Output.Choices)-1]
			// get last content text
			lastContentItem = lastChoice.Message.Content[len(lastChoice.Message.Content)-1]
			if lastContentItem.Text == "" {
				continue
			}
			lastCompleteDataLineText = lastContentItem.Text
			break
		}
	}
	if lastCompleteDataLineText == "" {
		return reverseproxy.Continue, nil
	}
	// 计算增量 Delta
	delta := strings.TrimPrefix(lastCompleteDataLineText, f.lastCompletionDataLineText)
	if delta == "" {
		return reverseproxy.Continue, nil
	}
	f.lastCompletionDataLineText = lastCompleteDataLineText
	// 更新 lastCompleteDataLine
	lastContentItem.Text = delta
	lastChoice.Message = QwenVLRespStreamChunkOutputChoiceMessage{Content: []QwenVLRespStreamChunkOutputChoiceMessageContent{lastContentItem}}
	lastCompleteDataLine.Output.Choices = []QwenVLRespStreamChunkOutputChoice{lastChoice}
	return reverseproxy.Continue, &lastCompleteDataLine
}

func TrimChunkDataPrefix(v []byte) []byte {
	s := strings.TrimPrefix(string(v), "data:")
	s = strings.TrimSpace(s)
	return []byte(s)
}

func convertToOpenAIStreamChunk(qwenVLResp QwenVLRespStreamChunk, modelName string) (*openai.ChatCompletionStreamResponse, error) {
	var ocs []openai.ChatCompletionStreamChoice
	for _, qwc := range qwenVLResp.Output.Choices {
		oc := openai.ChatCompletionStreamChoice{
			Index: 0,
			Delta: openai.ChatCompletionStreamChoiceDelta{
				Content: qwc.Message.Content[0].Text,
				Role:    qwc.Message.Role,
			},
			FinishReason: openai.FinishReason(qwc.FinishReason),
		}
		ocs = append(ocs, oc)
	}
	openaiChunk := openai.ChatCompletionStreamResponse{
		ID:      qwenVLResp.RequestID,
		Object:  "chat.completion.chunk",
		Model:   modelName,
		Choices: ocs,
	}
	return &openaiChunk, nil
}

func (f *DashScopeDirector) OnResponseEOF(ctx context.Context, _ reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (err error) {
	model := ctxhelper.MustGetModel(ctx)
	modelMeta := getModelMeta(model)
	modelName := modelMeta.Public.ModelName

	// only stream style need append [DONE] chunk
	if !ctxhelper.GetIsStream(ctx) {
		// convert all at once
		var qwVLChunk QwenVLRespStreamChunk
		if err := json.Unmarshal(f.DefaultResponseFilter.Buffer.Bytes(), &qwVLChunk); err != nil {
			return fmt.Errorf("failed to unmarshal qwen-vl response, chunk: %s, err: %v", string(chunk), err)
		}
		openaiResp, err := convertToOpenAIStreamChunk(qwVLChunk, string(modelName))
		if err != nil {
			return fmt.Errorf("failed to convert qwen-vl response to openai response, err: %v", err)
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
