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

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

type Filter struct {
	allChunks  []byte
	completion string
}

func init() {
	filter_define.RegisterFilterCreator("parse-openai-response", ResponseModifierCreator)
}

var (
	_ filter_define.ProxyResponseModifier = (*Filter)(nil)
)

var ResponseModifierCreator filter_define.ResponseModifierCreator = func(name string, config json.RawMessage) filter_define.ProxyResponseModifier {
	return &Filter{}
}

func (f *Filter) Enable(resp *http.Response) bool {
	auditRecID, ok := ctxhelper.GetAuditID(resp.Request.Context())
	return ok && auditRecID != ""
}

func (f *Filter) OnHeaders(resp *http.Response) error {
	return nil
}

func (f *Filter) OnBodyChunk(resp *http.Response, chunk []byte, index int64) (out []byte, err error) {
	f.allChunks = append(f.allChunks, chunk...)

	if ctxhelper.MustGetIsStream(resp.Request.Context()) {
		completion, _ := ExtractEventStreamCompletionAndFcName(string(chunk))
		f.completion += completion
	}

	return chunk, nil
}

func (f *Filter) OnComplete(resp *http.Response) (out []byte, err error) {
	if !ctxhelper.MustGetIsStream(resp.Request.Context()) {
		var completion string
		// routing by model type
		model := ctxhelper.MustGetModel(resp.Request.Context())
		switch model.Type {
		case modelpb.ModelType_text_generation, modelpb.ModelType_multimodal:
			completion, _ = ExtractApplicationJsonCompletionAndFcName(string(f.allChunks))
		case modelpb.ModelType_audio:
			var openaiAudioResp openai.AudioResponse
			if err := json.Unmarshal(f.allChunks, &openaiAudioResp); err == nil {
				completion = openaiAudioResp.Text
			} else {
				completion = string(f.allChunks)
			}
		case modelpb.ModelType_image:
			completion = string(f.allChunks)
			var openaiImageResp openai.ImageResponse
			if err := json.Unmarshal(f.allChunks, &openaiImageResp); err == nil {
				if len(openaiImageResp.Data) > 0 {
					completion = openaiImageResp.Data[0].URL
				}
			}
		}
		f.completion = completion
	}

	audithelper.Note(resp.Request.Context(), "completion", f.completion)

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
