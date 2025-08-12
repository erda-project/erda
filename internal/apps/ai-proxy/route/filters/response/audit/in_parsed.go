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

package audit

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func (f *AuditResponse) inParsedEnable(resp *http.Response) bool {
	auditRecID, ok := ctxhelper.GetAuditID(resp.Request.Context())
	return ok && auditRecID != ""
}

func (f *AuditResponse) inParsedOnHeaders(resp *http.Response) error {
	return nil
}

func (f *AuditResponse) inParsedOnBodyChunk(resp *http.Response, chunk []byte) (out []byte, err error) {
	if !f.inParsedEnable(resp) {
		return chunk, nil
	}
	f.inParsed.allChunks = append(f.inParsed.allChunks, chunk...)

	if ctxhelper.MustGetIsStream(resp.Request.Context()) {
		completion, responseFunctionCallName := ExtractEventStreamCompletionAndFcName(string(chunk))
		f.inParsed.completion += completion
		f.inParsed.responseFunctionCallName = f.inParsed.responseFunctionCallName + responseFunctionCallName + " "
	}

	return chunk, nil
}

func (f *AuditResponse) inParsedOnComplete(resp *http.Response) (out []byte, err error) {
	if !f.inParsedEnable(resp) {
		return nil, nil
	}

	if !ctxhelper.MustGetIsStream(resp.Request.Context()) {
		var completion, responseFunctionCallName string
		// routing by model type
		model := ctxhelper.MustGetModel(resp.Request.Context())
		switch model.Type {
		case modelpb.ModelType_text_generation, modelpb.ModelType_multimodal:
			completion, responseFunctionCallName = ExtractApplicationJsonCompletionAndFcName(string(f.inParsed.allChunks))
		case modelpb.ModelType_audio:
			var openaiAudioResp openai.AudioResponse
			if err := json.Unmarshal(f.inParsed.allChunks, &openaiAudioResp); err == nil {
				completion = openaiAudioResp.Text
			} else {
				completion = string(f.inParsed.allChunks)
			}
		case modelpb.ModelType_image:
			completion = string(f.inParsed.allChunks)
			var openaiImageResp openai.ImageResponse
			if err := json.Unmarshal(f.inParsed.allChunks, &openaiImageResp); err == nil {
				if len(openaiImageResp.Data) > 0 {
					completion = openaiImageResp.Data[0].URL
				}
			}
		}
		f.inParsed.completion = completion
		f.inParsed.responseFunctionCallName = responseFunctionCallName
	}

	auditRecID, _ := ctxhelper.GetAuditID(resp.Request.Context())
	// collect actual llm response info
	updateReq := pb.AuditUpdateRequestAfterLLMDirectorResponse{
		AuditId:    auditRecID,
		Completion: f.inParsed.completion,
		//ResponseBody: string(f.inParsed.allChunks), // TODO not store raw body anymore, but store parsed content
		ResponseHeader: func() string {
			b, err := json.Marshal(resp.Header)
			if err != nil {
				return err.Error()
			}
			return string(b)
		}(),
		ResponseFunctionCallName: f.inParsed.responseFunctionCallName,
	}

	// update audit into db
	dao := ctxhelper.MustGetDBClient(resp.Request.Context())
	dbClient := dao.AuditClient()
	_, err = dbClient.UpdateAfterLLMDirectorResponse(resp.Request.Context(), &updateReq)
	if err != nil {
		l := ctxhelper.MustGetLogger(resp.Request.Context())
		l.Errorf("failed to update audit on response EOF, audit id: %s, err: %v", auditRecID, err)
	}
	return nil, nil
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
