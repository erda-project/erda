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

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/dashscope-director/sdk"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func (f *DashScopeDirector) OnResponseChunk(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (signal reverseproxy.Signal, err error) {
	model := ctxhelper.MustGetModel(ctx)
	modelMeta := getModelMeta(model)
	responseType := mustGetResponseType(modelMeta)

	// switch by request type
	switch responseType {
	case metadata.AliyunDashScopeResponseTypeOpenAI:
		return f.DefaultResponseFilter.OnResponseChunk(ctx, infor, w, chunk) // return directly
	case metadata.AliyunDashScopeResponseTypeDs:
		return f.dsOnResponseChunk(ctx, infor, w, chunk)
	default:
		return reverseproxy.Intercept, fmt.Errorf("unsupported response type: %s", responseType)
	}
}

func (f *DashScopeDirector) dsOnResponseChunk(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (signal reverseproxy.Signal, err error) {
	if !ctxhelper.GetIsStream(ctx) {
		f.DefaultResponseFilter.Buffer.Write(chunk)
		return reverseproxy.Continue, nil
	}
	// handle chunk
	f.DefaultResponseFilter.Buffer.Write(chunk) // write to buffer, so we can get allChunks later
	return f.dsHandleResponseStreamChunk(ctx, w, chunk)
}

// dsHandleResponseStreamChunk
// example see ./resp_example.md
func (f *DashScopeDirector) dsHandleResponseStreamChunk(ctx context.Context, w reverseproxy.Writer, _ []byte) (reverseproxy.Signal, error) {
	model := ctxhelper.MustGetModel(ctx)
	modelMeta := getModelMeta(model)
	modelName := modelMeta.Public.ModelName

	// 因为 DashScope stream 不是真 chunk，每个 data line 的 text 都是在之前的基础上进行 append，所以需要主动计算增量值，模拟真 chunk 形式返回
	// 所以，只需要找到 chunk 里最后一个 data line，然后计算增量值，转换为一个 OpenAI Chunk 即可
	signal, lastCompleteDeltaResp := f.getDeltaTextFromStreamChunk(f.DefaultResponseFilter.Buffer.Bytes())
	if signal == reverseproxy.Intercept || lastCompleteDeltaResp == nil {
		return signal, nil
	}
	// convert ds response to openai response
	openaiChunk, err := sdk.ConvertDsStreamChunkToOpenAIStreamFormat(*lastCompleteDeltaResp, modelName)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to convert dashscope response to openai response, err: %v", err)
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

func (f *DashScopeDirector) getDeltaTextFromStreamChunk(allChunks []byte) (reverseproxy.Signal, *sdk.DsRespStreamChunk) {
	// 按行解析，保留 `data:` 开头的行
	var allDataLines []string
	scanner := bufio.NewScanner(bytes.NewReader(allChunks))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		allDataLines = append(allDataLines, line)
	}
	// try to get the last complete data line
	lastChunk, lastCompleteDataLineText, err := f.getLastChunk(allDataLines)
	if err != nil {
		return reverseproxy.Intercept, nil
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
	// construct new chunk
	newChunk := sdk.DsRespStreamChunk{
		Output: sdk.DsRespStreamChunkOutput{
			Choices: []sdk.DsRespStreamChunkOutputChoice{{Message: sdk.DsRespStreamChunkOutputChoiceMessage{Content: delta, Role: openai.ChatMessageRoleAssistant}}},
		},
		RequestID: lastChunk.RequestID,
	}
	return reverseproxy.Continue, &newChunk
}

func (f *DashScopeDirector) getLastChunk(allDataLines []string) (lastChunk *sdk.DsRespStreamChunk, content string, err error) {
	lastLine := allDataLines[len(allDataLines)-1]
	i := len(allDataLines) - 1
	if i < f.lastCompletionDataLineIndex {
		return
	}
	if err = json.Unmarshal(TrimChunkDataPrefix([]byte(lastLine)), &lastChunk); err != nil {
		return
	}
	f.lastCompletionDataLineIndex = i
	outputText := lastChunk.Output.Text
	if len(lastChunk.Output.Choices) == 0 && outputText == "" {
		return
	}
	// get text first
	if outputText != "" {
		content = outputText
		return
	}
	// get last choice
	lastChoice := &lastChunk.Output.Choices[len(lastChunk.Output.Choices)-1]
	// get last content text
	var lastContent string
	switch lastChoice.Message.Content.(type) {
	case string:
		lastContent = lastChoice.Message.Content.(string)
	case []any:
		var lastContentItems []sdk.DsRespStreamChunkOutputChoiceMessagePart
		cputil.MustObjJSONTransfer(lastChoice.Message.Content, &lastContentItems)
		lastContent = lastContentItems[len(lastContentItems)-1].Text
	default:
		err = fmt.Errorf("unsupported message content type: %T", lastChoice.Message.Content)
		return
	}
	content = handleOddContentStrIsJsonArray(lastContent)
	return
}

func TrimChunkDataPrefix(v []byte) []byte {
	s := strings.TrimPrefix(string(v), "data:")
	s = strings.TrimSpace(s)
	return []byte(s)
}

// handleOddContentStrIsJsonArray 处理 DashScope Kimi choice.message.content 是一个 string，但是 string 的内容是一个 json 数组字符串的情况
func handleOddContentStrIsJsonArray(lastContent string) string {
	hasPrefix := false
	strFlagToCheck := `[{"text":`
	// lastContent 为 strFlagToCheck 里的一部分，则等待
	if len(lastContent) <= len(strFlagToCheck) {
		if strings.HasPrefix(strFlagToCheck, lastContent) {
			// need wait
			return ""
		}
	}
	if strings.HasPrefix(lastContent, strFlagToCheck) {
		hasPrefix = true
		// remove prefix
		lastContent = strings.TrimPrefix(lastContent, strFlagToCheck)
		// find first `"`
		idx := strings.Index(lastContent, `"`)
		if idx < 0 {
			return lastContent
		}
		quotedContent := strings.TrimSpace(lastContent[idx+1:])
		lastContent = quotedContent
	}
	strSuccessNeedWait := []string{`"`, `"]`, `"]`}
	for _, suffix := range strSuccessNeedWait {
		if strings.HasSuffix(lastContent, suffix) {
			// need wait
			return ""
		}
	}
	// remove suffix
	if hasPrefix {
		lastContent = strings.TrimSuffix(lastContent, `"}]`)
	}
	return lastContent
}

// OnResponseEOF
// example see ./resp_example.md
func (f *DashScopeDirector) OnResponseEOF(ctx context.Context, _ reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (err error) {
	model := ctxhelper.MustGetModel(ctx)
	modelMeta := getModelMeta(model)
	modelName := modelMeta.Public.ModelName

	// only stream style need append [DONE] chunk
	if !ctxhelper.GetIsStream(ctx) {
		responseType := mustGetResponseType(modelMeta)
		if responseType == metadata.AliyunDashScopeResponseTypeOpenAI {
			return f.DefaultResponseFilter.OnResponseEOF(ctx, nil, w, chunk) // return directly
		}
		// convert all at once
		var dsResp sdk.DsResponse
		if err := json.Unmarshal(f.DefaultResponseFilter.Buffer.Bytes(), &dsResp); err != nil {
			return fmt.Errorf("failed to unmarshal dashscope stream chunk: %s, err: %v", string(chunk), err)
		}
		openaiResp, err := sdk.ConvertDsResponseToOpenAIFormat(dsResp, modelName)
		if err != nil {
			return fmt.Errorf("failed to convert dashscope chunk to openai format, err: %v", err)
		}
		b, err := json.Marshal(openaiResp)
		if err != nil {
			return fmt.Errorf("failed to marshal openai resp, err: %v", err)
		}
		return f.DefaultResponseFilter.OnResponseEOF(ctx, nil, w, b)
	}
	// append [DONE] chunk
	doneChunk := vars.ConcatChunkDataPrefix([]byte("[DONE]"))
	if _, err := w.Write(doneChunk); err != nil {
		return fmt.Errorf("failed to write openai chunk, err: %v", err)
	}
	return nil
}
