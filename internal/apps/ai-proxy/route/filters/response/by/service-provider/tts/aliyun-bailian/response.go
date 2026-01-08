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
	"fmt"
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/tts/ttsutil"
)

type (
	QwenTTSResponse struct {
		StatusCode int           `json:"status_code"`
		RequestID  string        `json:"request_id"`
		Code       string        `json:"code"`
		Message    string        `json:"message"`
		Output     QwenTTSOutput `json:"output"`
		Usage      QwenTTSUsage  `json:"usage"`
	}

	QwenTTSOutput struct {
		Text         *string      `json:"text"`
		FinishReason string       `json:"finish_reason"`
		Choices      []any        `json:"choices"`
		Audio        QwenTTSAudio `json:"audio"`
	}

	QwenTTSAudio struct {
		Data      string `json:"data"`
		URL       string `json:"url"`
		ID        string `json:"id"`
		ExpiresAt int64  `json:"expires_at"`
	}

	QwenTTSUsage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		Characters   int `json:"characters"`
	}
)

func (f *BailianTTSConverter) OnPeekChunkBeforeHeaders(resp *http.Response, peekBytes []byte) error {
	var qwenResp QwenTTSResponse
	if err := json.Unmarshal(peekBytes, &qwenResp); err != nil {
		return fmt.Errorf("failed to unmarshal qwen tts response: %w", err)
	}

	f.audioURL = qwenResp.Output.Audio.URL
	if f.audioURL == "" {
		return fmt.Errorf("missing audio url from qwen response: %s", f.audioURL)
	}

	// set content header to audio/xxx
	format, _ := ctxhelper.GetAudioTTSResponseFormat(resp.Request.Context())
	contentType := ttsutil.ContentTypeFromFormat(format)
	resp.Header.Set("Content-Type", contentType)

	return nil
}

func (f *BailianTTSConverter) OnBodyChunk(resp *http.Response, chunk []byte, index int64) (out []byte, err error) {
	// download audio and return
	audioBinary, err := ttsutil.DownloadAudio(resp.Request.Context(), f.audioURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio from qwen: %w", err)
	}
	return audioBinary, nil
}
