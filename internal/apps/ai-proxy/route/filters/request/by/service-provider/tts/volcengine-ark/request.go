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

package volcengine_ark

import (
	"encoding/json"
	"net/http/httputil"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/tts/ttsutil"
)

func (f *VolcengineTTSConverter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	// set x-api-request-id header
	pr.Out.Header.Set("x-api-request-id", ctxhelper.MustGetGeneratedCallID(pr.In.Context()))

	// handler request body
	var openaiReq openai.CreateSpeechRequest
	if err := json.NewDecoder(pr.Out.Body).Decode(&openaiReq); err != nil {
		return err
	}

	if openaiReq.ResponseFormat != "" {
		ctxhelper.PutAudioTTSResponseFormat(pr.In.Context(), string(openaiReq.ResponseFormat))
	}

	bytedanceReq := &BytedanceReq{
		ReqParams: ReqParams{
			Text:    openaiReq.Input,
			Speaker: ttsutil.MapVoice(openaiReq.Voice, FixedMaleVoice, FixedFemaleVoice),
			AudioParams: AudioParams{
				Format:     string(openai.SpeechResponseFormatMp3),
				SampleRate: 24000,
			},
		},
	}

	return body_util.SetBody(pr.Out, bytedanceReq)
}

type (
	BytedanceReq struct {
		ReqParams ReqParams `json:"req_params"`
	}
	ReqParams struct {
		Text        string      `json:"text"`
		Speaker     string      `json:"speaker"`
		AudioParams AudioParams `json:"audio_params"`
	}
	AudioParams struct {
		Format     string `json:"format"`
		SampleRate int    `json:"sample_rate"`
	}
)

const (
	FixedMaleVoice   = "zh_male_m191_uranus_bigtts"
	FixedFemaleVoice = "zh_female_vv_uranus_bigtts"
)
