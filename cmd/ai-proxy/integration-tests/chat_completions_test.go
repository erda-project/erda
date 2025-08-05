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
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/common"
	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/config"
)

// Use official type definitions from OpenAI Go library

func TestChatCompletionsNonStream(t *testing.T) {
	cfg := config.Get()
	client := common.NewClient()

	// Test messages
	testMessages := []openai.ChatCompletionMessage{
		{Role: "user", Content: "Hello, how are you?"},
	}

	// Get chat models for testing
	chatModels := cfg.ChatModels
	if len(chatModels) == 0 {
		t.Skip("No chat models configured for testing")
	}

	for _, model := range chatModels {
		t.Run(fmt.Sprintf("NonStream_Model_%s", model), func(t *testing.T) {
			testChatCompletionNonStreamForModel(t, client, model, testMessages)
		})
	}
}

func TestChatCompletionsStreaming(t *testing.T) {
	cfg := config.Get()
	client := common.NewClient()

	// Test messages (moderate length prompt to test true streaming response)
	testMessages := []openai.ChatCompletionMessage{
		{Role: "user", Content: "Write a 100-word introduction about Hangzhou"},
	}

	// Get chat models for testing
	chatModels := cfg.ChatModels
	if len(chatModels) == 0 {
		t.Skip("No chat models configured for testing")
	}

	for _, model := range chatModels {
		t.Run(fmt.Sprintf("Streaming_Model_%s", model), func(t *testing.T) {
			testChatCompletionStreamingForModel(t, client, model, testMessages)
		})
	}
}

func testChatCompletionNonStreamForModel(t *testing.T, client *common.Client, model string, messages []openai.ChatCompletionMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	request := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   100,
		Temperature: 0.7,
		Stream:      false,
	}

	startTime := time.Now()
	resp := client.PostJSON(ctx, "/v1/chat/completions", request)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Request failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Request failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var chatResp openai.ChatCompletionResponse
	if err := resp.GetJSON(&chatResp); err != nil {
		t.Fatalf("✗ Failed to parse response: %v", err)
	}

	// Validate response
	if chatResp.ID == "" {
		t.Error("✗ Response ID is empty")
	}
	if chatResp.Object != "chat.completion" {
		t.Errorf("✗ Expected object 'chat.completion', got '%s'", chatResp.Object)
	}
	if len(chatResp.Choices) == 0 {
		t.Error("✗ No choices in response")
	}
	if chatResp.Choices[0].Message.Content == "" {
		t.Error("✗ Empty message content")
	}
	if chatResp.Choices[0].FinishReason == "" {
		t.Error("✗ Empty finish reason")
	}

	t.Logf("✓ Non-stream Model %s: %s (response time: %v)", model, strings.TrimSpace(chatResp.Choices[0].Message.Content), responseTime)
}

func testChatCompletionStreamingForModel(t *testing.T, client *common.Client, model string, messages []openai.ChatCompletionMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	request := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   100, // Moderate token count
		Temperature: 0.7,
		Stream:      true,
	}

	var content strings.Builder
	var streamCount int
	var chunkTimes []time.Time // Record arrival time of each chunk

	resp := client.PostJSONStream(ctx, "/v1/chat/completions", request, func(data []byte) error {
		chunkTimes = append(chunkTimes, time.Now()) // Record chunk arrival time
		var streamResp openai.ChatCompletionStreamResponse
		if err := json.Unmarshal(data, &streamResp); err != nil {
			return fmt.Errorf("failed to parse stream response: %w", err)
		}

		streamCount++

		// Validate streaming response
		// When choices length is 0, allow other fields to be empty, e.g., first chunk of gpt-4o only has prompt_filter_results field with value
		if len(streamResp.Choices) == 0 && len(streamResp.PromptFilterResults) == 0 && streamResp.Usage == nil {
			return fmt.Errorf("stream response has no choices, prompt_filter_results, or usage")
		}
		if len(streamResp.Choices) > 0 {
			if len(streamResp.ID) == 0 {
				return fmt.Errorf("stream response ID is empty")
			}
			if streamResp.Object != "chat.completion.chunk" {
				return fmt.Errorf("expected object 'chat.completion.chunk', got '%s'", streamResp.Object)
			}
			// Model name may change, only do basic validation
			if streamResp.Model == "" {
				return fmt.Errorf("stream response model is empty")
			}
			// Accumulate content
			content.WriteString(streamResp.Choices[0].Delta.Content)
		}

		return nil
	})

	if resp.Error != nil {
		t.Fatalf("✗ Streaming request failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Streaming request failed with status %d", resp.StatusCode)
	}

	// Check if response header is text/event-stream
	contentType := resp.Headers.Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("✗ Expected Content-Type: text/event-stream, got: %s", contentType)
	}

	if streamCount == 0 {
		t.Error("✗ No stream responses received")
	}

	if content.Len() == 0 {
		t.Error("✗ No content received from stream")
	}

	// Analyze chunk arrival time to detect if it's true streaming
	analyzeStreamingTiming(t, chunkTimes, model)

	contentStr := strings.TrimSpace(content.String())
	t.Logf("✓ Streaming Model %s: %d chunks", model, streamCount)
	t.Logf("  Content preview: %s...", truncateString(contentStr, 100))
}

// analyzeStreamingTiming analyzes timing characteristics of streaming
func analyzeStreamingTiming(t *testing.T, chunkTimes []time.Time, model string) {
	if len(chunkTimes) < 2 {
		t.Logf("⚠ Model %s: Too few chunks to analyze streaming timing (%d chunks)", model, len(chunkTimes))
		return
	}

	// Calculate time intervals
	var intervals []time.Duration
	for i := 1; i < len(chunkTimes); i++ {
		interval := chunkTimes[i].Sub(chunkTimes[i-1])
		intervals = append(intervals, interval)
	}

	// Calculate statistics
	totalDuration := chunkTimes[len(chunkTimes)-1].Sub(chunkTimes[0])
	avgInterval := totalDuration / time.Duration(len(intervals))

	// Check if all chunks arrive almost simultaneously (fake streaming)
	const maxFakeStreamThreshold = 1 * time.Millisecond // Consider as simultaneous arrival within 1ms
	isFakeStream := true
	for _, interval := range intervals {
		if interval > maxFakeStreamThreshold {
			isFakeStream = false
			break
		}
	}

	t.Logf("📊 Streaming timing analysis for model %s:", model)
	t.Logf("   Total chunks: %d", len(chunkTimes))
	t.Logf("   Total duration: %v", totalDuration)
	t.Logf("   Average interval: %v", avgInterval)

	if isFakeStream {
		t.Logf("   ⚠ DETECTED: Fake streaming (Maybe WholeStreamSplitter) - all chunks arrived within %v", maxFakeStreamThreshold)
		t.Logf("   This indicates the response was split from a complete response, not true streaming")
		t.Errorf("✗ STREAMING TEST FAILED for model %s: Detected fake streaming (WholeStreamSplitter). Expected true streaming but got batch-split response.", model)
	} else {
		t.Logf("   ✓ DETECTED: True streaming - chunks arrived with meaningful intervals")
	}

	// Show first few intervals as examples (in milliseconds)
	maxShow := 5
	if len(intervals) < maxShow {
		maxShow = len(intervals)
	}
	intervalsMs := make([]string, maxShow)
	for i := 0; i < maxShow; i++ {
		intervalsMs[i] = fmt.Sprintf("%.3fms", float64(intervals[i].Nanoseconds())/1000000.0)
	}
	t.Logf("   First %d intervals: [%s]", maxShow, strings.Join(intervalsMs, " "))
}

// truncateString truncates string to specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
