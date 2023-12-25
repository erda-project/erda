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

package audit_before_llm_director

import (
	"bufio"
	"context"
	"encoding/json"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func (f *Filter) OnResponseChunkImmutable(ctx context.Context, infor reverseproxy.HttpInfor, copiedChunk []byte) (signal reverseproxy.Signal, err error) {
	return reverseproxy.Continue, nil
}

func (f *Filter) OnResponseEOFImmutable(ctx context.Context, infor reverseproxy.HttpInfor, copiedChunk []byte) (err error) {
	auditRecID, ok := ctxhelper.GetAuditID(ctx)
	if !ok || auditRecID == "" {
		return nil
	}
	respBuffer := ctxhelper.GetLLMDirectorActualResponseBuffer(ctx)
	var completion, responseFunctionCallName string

	// routing by model type
	model, _ := ctxhelper.GetModel(ctx)
	switch model.Type {
	case modelpb.ModelType_text_generation:
		if ctxhelper.GetIsStream(ctx) {
			completion, responseFunctionCallName = ExtractEventStreamCompletionAndFcName(respBuffer.String())
		} else {
			completion, responseFunctionCallName = ExtractApplicationJsonCompletionAndFcName(respBuffer.String())
		}
	case modelpb.ModelType_audio:
		var openaiAudioResp openai.AudioResponse
		if err := json.Unmarshal([]byte(respBuffer.String()), &openaiAudioResp); err == nil {
			completion = openaiAudioResp.Text
		} else {
			completion = respBuffer.String()
		}
	case modelpb.ModelType_image:
		completion = respBuffer.String()
		var openaiImageResp openai.ImageResponse
		if err := json.Unmarshal([]byte(respBuffer.String()), &openaiImageResp); err == nil {
			if len(openaiImageResp.Data) > 0 {
				completion = openaiImageResp.Data[0].URL
			}
		}
	}

	// collect actual llm response info
	updateReq := pb.AuditUpdateRequestAfterLLMDirectorResponse{
		AuditId:      auditRecID,
		Completion:   completion,
		ResponseBody: respBuffer.String(),
		ResponseHeader: func() string {
			b, err := json.Marshal(infor.Header())
			if err != nil {
				return err.Error()
			}
			return string(b)
		}(),
		ResponseFunctionCallName: responseFunctionCallName,
	}

	// update audit into db
	dao := ctxhelper.MustGetDBClient(ctx)
	dbClient := dao.AuditClient()
	_, err = dbClient.UpdateAfterLLMDirectorResponse(ctx, &updateReq)
	if err != nil {
		l := ctxhelper.GetLogger(ctx)
		l.Errorf("failed to update audit on response EOF, audit id: %s, err: %v", auditRecID, err)
	}
	return nil
}

func ExtractEventStreamCompletionAndFcName(responseBody string) (completion string, fcName string) {
	scanner := bufio.NewScanner(strings.NewReader(responseBody))
	for scanner.Scan() {
		var line = scanner.Text()
		left := strings.Index(line, "{")
		right := strings.LastIndex(line, "}")
		if left < 0 || right < 1 {
			continue
		}
		line = line[left : right+1]

		var m = make(map[string]json.RawMessage)
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		choices, ok := m["choices"]
		if !ok {
			continue
		}
		var items []openai.ChatCompletionStreamChoice
		if err := json.Unmarshal(choices, &items); err != nil {
			continue
		}
		if len(items) == 0 {
			continue
		}
		delta := items[len(items)-1].Delta
		completion += delta.Content
		if delta.FunctionCall != nil {
			if delta.FunctionCall.Name != "" {
				fcName = delta.FunctionCall.Name
			}
			completion += delta.FunctionCall.Arguments
		}
	}

	return completion, fcName
}

func ExtractApplicationJsonCompletionAndFcName(responseBody string) (completion string, fcName string) {
	var m = make(map[string]json.RawMessage)
	if err := json.NewDecoder(strings.NewReader(responseBody)).Decode(&m); err != nil {
		return
	}
	data, ok := m["choices"]
	if !ok {
		return
	}
	var choices []*openai.ChatCompletionChoice
	if err := json.Unmarshal(data, &choices); err != nil {
		return
	}
	if len(choices) == 0 {
		return
	}
	msg := choices[0].Message
	completion = msg.Content
	if msg.FunctionCall != nil {
		fcName = msg.FunctionCall.Name
		completion = msg.FunctionCall.Arguments
	}
	return
}
