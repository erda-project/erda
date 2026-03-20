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

package message_converter

import "testing"

func TestAnthropicResponseConvertToOpenAIFormat_MapsCacheReadIntoPromptDetails(t *testing.T) {
	resp := AnthropicResponse{
		ID:   "msg_1",
		Role: "assistant",
		Type: "message",
		Content: []AnthropicResponseContentPart{
			{"type": "text", "text": "hello"},
		},
		Usage: AnthropicResponseUsage{
			InputTokens:              50,
			CacheReadInputTokens:     100,
			CacheCreationInputTokens: 25,
			OutputTokens:             10,
		},
	}

	openaiResp, err := resp.ConvertToOpenAIFormat("claude-sonnet")
	if err != nil {
		t.Fatalf("ConvertToOpenAIFormat unexpected error: %v", err)
	}

	if got := openaiResp.Usage.PromptTokens; got != 175 {
		t.Fatalf("expected prompt tokens 175, got %d", got)
	}
	if openaiResp.Usage.PromptTokensDetails == nil {
		t.Fatalf("expected prompt token details to be populated")
	}
	if got := openaiResp.Usage.PromptTokensDetails.CachedTokens; got != 100 {
		t.Fatalf("expected cached tokens 100, got %d", got)
	}
	if got := openaiResp.Usage.TotalTokens; got != 185 {
		t.Fatalf("expected total tokens 185, got %d", got)
	}
}

func TestConvertStreamChunkDataToOpenAIChunk_MessageStopMapsCacheReadIntoPromptDetails(t *testing.T) {
	input := AnthropicStreamMessageInfo{
		ID:    "msg_1",
		Model: "claude-sonnet",
		Role:  "assistant",
		Usage: AnthropicResponseUsage{
			InputTokens:              50,
			CacheReadInputTokens:     100,
			CacheCreationInputTokens: 25,
			OutputTokens:             10,
		},
	}

	_, chunk, err := ConvertStreamChunkDataToOpenAIChunk([]byte(`{"type":"message_stop"}`), input)
	if err != nil {
		t.Fatalf("ConvertStreamChunkDataToOpenAIChunk unexpected error: %v", err)
	}
	if chunk == nil || chunk.Usage == nil {
		t.Fatalf("expected usage chunk to be returned")
	}
	if got := chunk.Usage.PromptTokens; got != 175 {
		t.Fatalf("expected prompt tokens 175, got %d", got)
	}
	if chunk.Usage.PromptTokensDetails == nil {
		t.Fatalf("expected prompt token details to be populated")
	}
	if got := chunk.Usage.PromptTokensDetails.CachedTokens; got != 100 {
		t.Fatalf("expected cached tokens 100, got %d", got)
	}
	if got := chunk.Usage.TotalTokens; got != 185 {
		t.Fatalf("expected total tokens 185, got %d", got)
	}
}
