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

package impl

import (
	"strings"
	"testing"
)

func TestParseOpenAIStreamUsage(t *testing.T) {
	body := []byte("data: {\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":2,\"total_tokens\":5}}\n" +
		"data: [DONE]\n")

	usage := parseOpenAIStreamUsage(body)

	if usage == nil {
		t.Fatalf("expected usage map, got nil")
	}
	if v, ok := getUintFromMaps(usage, "input_tokens", "prompt_tokens"); !ok || v != 3 {
		t.Fatalf("expected prompt_tokens=3, got %d (ok=%v)", v, ok)
	}
	if v, ok := getUintFromMaps(usage, "output_tokens", "completion_tokens"); !ok || v != 2 {
		t.Fatalf("expected completion_tokens=2, got %d (ok=%v)", v, ok)
	}
	if v, ok := getUintFromMaps(usage, "total_tokens"); !ok || v != 5 {
		t.Fatalf("expected total_tokens=5, got %d (ok=%v)", v, ok)
	}

}

func TestParseOpenAIStreamUsageWithDetailedChunk(t *testing.T) {
	body := []byte("data: {\"choices\":[],\"created\":1760345759,\"id\":\"chatcmpl-CQ8dTTi66IblzcCIERJkis7P6Gx9h\",\"model\":\"gpt-5-2025-08-07\",\"obfuscation\":\"\",\"object\":\"chat.completion.chunk\",\"system_fingerprint\":null,\"usage\":{\"completion_tokens\":782,\"completion_tokens_details\":{\"accepted_prediction_tokens\":0,\"audio_tokens\":0,\"reasoning_tokens\":640,\"rejected_prediction_tokens\":0},\"prompt_tokens\":15,\"prompt_tokens_details\":{\"audio_tokens\":0,\"cached_tokens\":0},\"total_tokens\":797}}\n")

	usage := parseOpenAIStreamUsage(body)

	if usage == nil {
		t.Fatalf("expected usage map, got nil")
	}
	if v, ok := getUintFromMaps(usage, "total_tokens"); !ok || v != 797 {
		t.Fatalf("expected total_tokens=797, got %d (ok=%v)", v, ok)
	}
	if v, ok := getUintFromMaps(usage, "prompt_tokens"); !ok || v != 15 {
		t.Fatalf("expected prompt_tokens=15, got %d (ok=%v)", v, ok)
	}
	if v, ok := getUintFromMaps(usage, "completion_tokens", "output_tokens"); !ok || v != 782 {
		t.Fatalf("expected completion_tokens=782, got %d (ok=%v)", v, ok)
	}
}

func TestParseOpenAIStreamUsageFallbackToFullBody(t *testing.T) {
	usageLine := []byte("data: {\"usage\":{\"prompt_tokens\":1,\"total_tokens\":2}}\n")
	largeChunk := "data: {\"output\":\"" + strings.Repeat("a", 2500) + "\"}\n"
	body := append([]byte{}, usageLine...)
	body = append(body, largeChunk...)
	body = append(body, []byte("data: [DONE]\n")...)

	usage := parseOpenAIStreamUsage(body)

	if usage == nil {
		t.Fatalf("expected usage map, got nil")
	}
	if v, ok := getUintFromMaps(usage, "prompt_tokens"); !ok || v != 1 {
		t.Fatalf("expected prompt_tokens=1, got %d (ok=%v)", v, ok)
	}
	if v, ok := getUintFromMaps(usage, "total_tokens"); !ok || v != 2 {
		t.Fatalf("expected total_tokens=2, got %d (ok=%v)", v, ok)
	}
}

func TestParseOpenAIStreamUsageExtractsUsageValues(t *testing.T) {
	body := []byte("event: response.completed\n" +
		"data: {\"type\":\"response.completed\",\"usage\":{\"input_tokens\":91,\"output_tokens\":1461,\"total_tokens\":1552}}\n" +
		"data: [DONE]\n")

	usage := parseOpenAIStreamUsage(body)

	if usage == nil {
		t.Fatalf("expected usage map, got nil")
	}
	if v, ok := getUintFromMaps(usage, "input_tokens", "prompt_tokens"); !ok || v != 91 {
		t.Fatalf("expected input_tokens=91, got %d (ok=%v)", v, ok)
	}
	if v, ok := getUintFromMaps(usage, "output_tokens", "completion_tokens"); !ok || v != 1461 {
		t.Fatalf("expected output_tokens=1461, got %d (ok=%v)", v, ok)
	}
	if v, ok := getUintFromMaps(usage, "total_tokens"); !ok || v != 1552 {
		t.Fatalf("expected total_tokens=1552, got %d (ok=%v)", v, ok)
	}
}
