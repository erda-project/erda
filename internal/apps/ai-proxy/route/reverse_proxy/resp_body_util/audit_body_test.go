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

package resp_body_util

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTruncateBodyForAudit(t *testing.T) {
	body := []byte("abcdefghijklmnopqrstuvwxyz")
	got := TruncateBodyForAudit(body, 5, 3)
	want := "abcde...[omitted 18 bytes]...xyz"
	if string(got) != want {
		t.Fatalf("unexpected truncation result: %q", string(got))
	}
}

// makeSSEBody joins SSE events with the standard \n\n delimiter.
func makeSSEBody(events []string) string {
	return strings.Join(events, "\n\n") + "\n\n"
}

func TestOptimizeBodyForAudit(t *testing.T) {
	longDesc := strings.Repeat("x", 1000)
	tool := map[string]interface{}{
		"type":        "function",
		"name":        "my_tool",
		"description": longDesc,
		"parameters":  map[string]interface{}{"type": "object"},
		"strict":      true,
	}
	response := map[string]interface{}{"id": "r1", "output": []interface{}{}, "tools": []interface{}{tool}}
	createdEvent, _ := json.Marshal(map[string]interface{}{"type": "response.created", "response": response})

	tests := []struct {
		name        string
		body        string
		headLimit   int
		tailLimit   int
		contains    []string
		notContains []string
	}{
		{
			name:      "non-SSE falls back to truncation",
			body:      `{"choices":[{"message":{"content":"hello"}}]}`,
			headLimit: 5,
			tailLimit: 3,
			contains:  []string{"omitted"},
		},
		{
			name:      "json body containing data colon is not treated as SSE",
			body:      `{"data":"value","message":"` + strings.Repeat("x", 64) + `"}`,
			headLimit: 12,
			tailLimit: 8,
			contains:  []string{"omitted"},
		},
		{
			name: "delta events are dropped",
			body: makeSSEBody([]string{
				"event: response.created\ndata: " + `{"type":"response.created","response":{"id":"r1","output":[],"tools":[]}}`,
				"event: response.text.delta\ndata: " + `{"type":"response.text.delta","delta":"Hello"}`,
				"event: response.output_text.delta\ndata: " + `{"type":"response.output_text.delta","delta":" world"}`,
				"event: response.done\ndata: " + `{"type":"response.done","response":{"id":"r1","output":[{"type":"message","content":[{"type":"output_text","text":"Hello world"}]}]}}`,
			}),
			headLimit:   1024 * 30,
			tailLimit:   1024 * 2,
			contains:    []string{"response.created", "response.done"},
			notContains: []string{"response.text.delta", "response.output_text.delta"},
		},
		{
			name:        "tool schemas are compressed",
			body:        makeSSEBody([]string{"event: response.created\ndata: " + string(createdEvent)}),
			headLimit:   1024 * 30,
			tailLimit:   1024 * 2,
			contains:    []string{"my_tool"},
			notContains: []string{longDesc, `"parameters"`, `"strict"`},
		},
		{
			name: "size cap applied after SSE optimization",
			body: makeSSEBody([]string{
				"event: response.done\ndata: " + `{"type":"response.done","response":{"id":"r1","output":[{"type":"message","content":[{"type":"output_text","text":"` + strings.Repeat("a", 40000) + `"}]}]}}`,
			}),
			headLimit: 10,
			tailLimit: 5,
			contains:  []string{"omitted"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(OptimizeBodyForAudit([]byte(tt.body), tt.headLimit, tt.tailLimit))
			if tt.name == "json body containing data colon is not treated as SSE" {
				want := string(TruncateBodyForAudit([]byte(tt.body), tt.headLimit, tt.tailLimit))
				if result != want {
					t.Fatalf("expected JSON body to fall back to truncation, got %q want %q", result, want)
				}
			}
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("expected result to contain %q, got: %s", s, result)
				}
			}
			for _, s := range tt.notContains {
				if strings.Contains(result, s) {
					t.Errorf("expected result NOT to contain %q", s)
				}
			}
		})
	}
}
