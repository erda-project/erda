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

package qwen

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type QwenTTSResponse struct {
	StatusCode int           `json:"status_code"`
	RequestID  string        `json:"request_id"`
	Code       string        `json:"code"`
	Message    string        `json:"message"`
	Output     QwenTTSOutput `json:"output"`
	Usage      QwenTTSUsage  `json:"usage"`
}

type QwenTTSOutput struct {
	Text         *string      `json:"text"`
	FinishReason string       `json:"finish_reason"`
	Choices      []any        `json:"choices"`
	Audio        QwenTTSAudio `json:"audio"`
}

type QwenTTSAudio struct {
	Data      string `json:"data"`
	URL       string `json:"url"`
	ID        string `json:"id"`
	ExpiresAt int64  `json:"expires_at"`
}

type QwenTTSUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	Characters   int `json:"characters"`
}

func (f *QwenTTSConverter) OnHeaders(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}
	resp.Header.Set("Content-Type", "audio/mpeg")
	return nil
}

func (f *QwenTTSConverter) OnBodyChunk(resp *http.Response, chunk []byte, index int64) (out []byte, err error) {
	f.buff.Write(chunk)
	return nil, nil
}

func (f *QwenTTSConverter) OnComplete(resp *http.Response) (out []byte, err error) {
	// 1. parse qwen response
	var qwenResp QwenTTSResponse
	if err := json.NewDecoder(&f.buff).Decode(&qwenResp); err != nil {
		return out, fmt.Errorf("failed to parse qwen response: %w", err)
	}

	// 2. get audio url
	if qwenResp.Output.Audio.URL == "" {
		return out, fmt.Errorf("missing audio url from qwen response")
	}

	// 3. download audio
	audioBinary, err := f.downloadAudio(resp.Request.Context(), qwenResp.Output.Audio.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio: %w", err)
	}

	return audioBinary, nil
}

func (f *QwenTTSConverter) downloadAudio(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
