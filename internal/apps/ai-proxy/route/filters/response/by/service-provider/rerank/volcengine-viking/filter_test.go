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

package volcengine_viking

import (
	"net/http"
	"strings"
	"testing"
)

func TestOnBodyChunk(t *testing.T) {
	converter := &ResponseConverter{}
	resp := &http.Response{Request: &http.Request{}}

	chunk := []byte(`{"code":0,"msg":"ok","data":{"scores":[0.2,0.9,0.5],"token_usage":{"total_tokens":12}}}`)
	out, err := converter.OnBodyChunk(resp, chunk, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := string(out)
	if got == string(chunk) {
		t.Fatalf("expected converted output, got original chunk")
	}
	if !contains(got, `"relevance_score":0.9`) || !contains(got, `"index":1`) {
		t.Fatalf("converted output missing top score/index, got: %s", got)
	}
	if !contains(got, `"total_tokens":12`) {
		t.Fatalf("converted output missing token usage, got: %s", got)
	}
}

func TestExtractTotalTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int64
	}{
		{name: "float", input: float64(9), expected: 9},
		{name: "int64", input: int64(11), expected: 11},
		{name: "snake_case", input: map[string]any{"total_tokens": float64(15)}, expected: 15},
		{name: "camel_case", input: map[string]any{"totalTokens": float64(17)}, expected: 17},
		{name: "unknown", input: map[string]any{"x": 1}, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTotalTokens(tt.input)
			if got != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }
