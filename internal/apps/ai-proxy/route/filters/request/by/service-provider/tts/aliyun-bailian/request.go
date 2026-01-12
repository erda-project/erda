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

package aliyun_bailian

import (
	"encoding/json"
	"net/http/httputil"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
)

func (f *BailianTTSConverter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	var openaiReq openai.CreateSpeechRequest
	if err := json.NewDecoder(pr.Out.Body).Decode(&openaiReq); err != nil {
		return err
	}

	if openaiReq.ResponseFormat != "" {
		ctxhelper.PutAudioTTSResponseFormat(pr.In.Context(), string(openaiReq.ResponseFormat))
	}

	qwenReq := &QwenReq{
		Model: string(openaiReq.Model),
		Input: QwenInput{
			Text:  openaiReq.Input,
			Voice: mapVoice(openaiReq.Voice),
		},
	}

	if openaiReq.Voice != "" {
		qwenReq.Input.Voice = mapVoice(openaiReq.Voice)
	}

	return body_util.SetBody(pr.Out, qwenReq)
}

type QwenReq struct {
	Model string    `json:"model"`
	Input QwenInput `json:"input"`
}

type QwenInput struct {
	Text         string `json:"text"`
	Voice        string `json:"voice,omitempty"`
	LanguageType string `json:"language_type,omitempty"`
}

const (
	FixedMaleVoice   = "Ethan"
	FixedFemaleVoice = "Cherry"
)

var genderMap = map[openai.SpeechVoice]string{
	// female
	openai.VoiceAlloy:   "female",
	openai.VoiceNova:    "female",
	openai.VoiceShimmer: "female",
	"woman":             "female",
	"female":            "female",

	// male
	openai.VoiceEcho:  "male",
	openai.VoiceFable: "male",
	openai.VoiceOnyx:  "male",
	"man":             "male",
	"male":            "male",
}

func mapVoice(voice openai.SpeechVoice) string {
	gender := genderMap[voice]

	switch gender {
	case "female":
		return FixedFemaleVoice
	case "male":
		return FixedMaleVoice
	default:
		// return original if mismatch
		return string(voice)
	}
}
