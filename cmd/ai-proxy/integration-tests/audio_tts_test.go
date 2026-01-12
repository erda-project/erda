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

package integration_tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/common"
	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/config"
)

// AudioSpeechRequest represents the TTS request structure
type AudioSpeechRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice,omitempty"`
	ResponseFormat string  `json:"response_format,omitempty"` // mp3, opus, aac, flac, wav, pcm
	Speed          float64 `json:"speed,omitempty"`
}

func TestAudioSpeech(t *testing.T) {
	cfg := config.Get()
	client := common.NewClient()

	ttsModels := cfg.AudioTTSModels
	if len(ttsModels) == 0 {
		t.Skip("No TTS models configured for testing (AUDIO_TTS_MODELS)")
	}

	testInputs := []struct {
		input          string
		responseFormat string
	}{
		{"Hello world, this is a TTS test.", "mp3"},
		{"你好世界，这是一个 TTS 测试。", "wav"},
	}

	for _, model := range ttsModels {
		for _, test := range testInputs {
			t.Run(fmt.Sprintf("Model_%s_Format_%s", model, test.responseFormat), func(t *testing.T) {
				testTTSForModel(t, client, model, test.input, test.responseFormat)
			})
		}
	}
}

func testTTSForModel(t *testing.T, client *common.Client, model, input, format string) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	req := AudioSpeechRequest{
		Model:          model,
		Input:          input,
		ResponseFormat: format,
	}
	modelLower := strings.ToLower(model)
	if strings.Contains(modelLower, "qwen") || strings.Contains(modelLower, "bailian") ||
		strings.Contains(modelLower, "doubao") || strings.Contains(modelLower, "bytedance") {
		req.Voice = "woman"
	}

	startTime := time.Now()
	resp := client.PostJSON(ctx, "/v1/audio/speech", req)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Request failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Request failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// Validate Content-Type
	contentType := resp.Headers.Get("Content-Type")
	expectedContentType := ""
	switch format {
	case "mp3":
		expectedContentType = "audio/mpeg"
	case "wav":
		expectedContentType = "audio/wav"
	default:
		expectedContentType = "audio/"
	}

	if !strings.Contains(contentType, expectedContentType) {
		t.Errorf("✗ Unexpected Content-Type: expected to contain %s, got %s", expectedContentType, contentType)
	}

	// Validate Body
	if len(resp.Body) < 100 { // Audio file should at least have some bytes
		t.Errorf("✗ Response body too small: %d bytes", len(resp.Body))
	}

	t.Logf("✓ Model %s generated %s audio: %d bytes (response time: %v)",
		model, format, len(resp.Body), responseTime)
}

func TestAudioSpeechErrorHandling(t *testing.T) {
	cfg := config.Get()
	client := common.NewClient()

	ttsModels := cfg.AudioTTSModels
	if len(ttsModels) == 0 {
		t.Skip("No TTS models configured for testing")
	}

	model := ttsModels[0]

	t.Run("EmptyInput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
		defer cancel()

		req := AudioSpeechRequest{
			Model: model,
			Input: "",
		}
		resp := client.PostJSON(ctx, "/v1/audio/speech", req)

		if resp.IsSuccess() {
			t.Error("✗ Expected request to fail with empty input, but it succeeded")
		} else {
			t.Logf("✓ Correctly rejected empty input (status: %d)", resp.StatusCode)
		}
	})

	t.Run("InvalidModel", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
		defer cancel()

		req := AudioSpeechRequest{
			Model: "non-existent-model",
			Input: "Hello",
		}
		resp := client.PostJSON(ctx, "/v1/audio/speech", req)

		if resp.IsSuccess() {
			t.Error("✗ Expected request to fail with invalid model, but it succeeded")
		} else {
			t.Logf("✓ Correctly rejected invalid model (status: %d)", resp.StatusCode)
		}
	})
}
