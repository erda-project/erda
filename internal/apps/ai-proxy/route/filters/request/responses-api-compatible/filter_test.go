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

package responses_api_compatible

import (
	"reflect"
	"testing"
)

func TestNormalizeInputMessages(t *testing.T) {
	tests := []struct {
		name          string
		input         any
		expected      any
		expectChanged bool
	}{
		{
			name:  "string input becomes user message",
			input: "hello world",
			expected: []any{
				map[string]any{
					"role": "user",
					"content": []any{
						map[string]any{
							"type": inputContentTypeInputText,
							"text": "hello world",
						},
					},
				},
			},
			expectChanged: true,
		},
		{
			name: "single message map normalized",
			input: map[string]any{
				"type":        "message",
				"id":          "msg-1",
				"role":        "assistant",
				"content":     "ping",
				"annotations": "should-remove",
				"logprobs":    "drop-me",
			},
			expected: []any{
				map[string]any{
					"type": "message",
					"role": "user",
					"content": []any{
						map[string]any{
							"type": inputContentTypeInputText,
							"text": "ping",
						},
					},
				},
			},
			expectChanged: true,
		},
		{
			name: "function call kept intact except metadata",
			input: []any{
				map[string]any{
					"type":        "function_call",
					"id":          "call-1",
					"name":        "get_weather",
					"arguments":   map[string]any{"city": "Shanghai"},
					"annotations": "to-remove",
					"logprobs":    []any{"drop"},
					"content": []any{
						map[string]any{
							"type": "json_schema",
							"data": "original",
						},
					},
				},
			},
			expected: []any{
				map[string]any{
					"type":      "function_call",
					"name":      "get_weather",
					"arguments": map[string]any{"city": "Shanghai"},
					"content": []any{
						map[string]any{
							"type": "json_schema",
							"data": "original",
						},
					},
				},
			},
			expectChanged: true,
		},
		{
			name: "message content array normalized",
			input: []any{
				map[string]any{
					"type": "message",
					"role": "user",
					"content": []any{
						map[string]any{"type": "plain_text", "text": "alpha"},
						"beta",
					},
				},
			},
			expected: []any{
				map[string]any{
					"type": "message",
					"role": "user",
					"content": []any{
						map[string]any{"type": inputContentTypeInputText, "text": "alpha"},
						map[string]any{"type": inputContentTypeInputText, "text": "beta"},
					},
				},
			},
			expectChanged: true,
		},
		{
			name: "message without content gets default part",
			input: []any{
				map[string]any{
					"type": "message",
					"role": "user",
				},
			},
			expected: []any{
				map[string]any{
					"type": "message",
					"role": "user",
					"content": []any{
						map[string]any{
							"type": inputContentTypeInputText,
							"text": "",
						},
					},
				},
			},
			expectChanged: true,
		},
		{
			name: "already normalized message stays unchanged",
			input: []any{
				map[string]any{
					"type": "message",
					"role": "user",
					"content": []any{
						map[string]any{
							"type": inputContentTypeInputText,
							"text": "ready",
						},
					},
				},
			},
			expected: []any{
				map[string]any{
					"type": "message",
					"role": "user",
					"content": []any{
						map[string]any{
							"type": inputContentTypeInputText,
							"text": "ready",
						},
					},
				},
			},
			expectChanged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed := normalizeInputMessages(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("unexpected normalized result:\nexpected: %#v\nactual:   %#v", tt.expected, got)
			}
			if changed != tt.expectChanged {
				t.Fatalf("expected changed=%v, got %v", tt.expectChanged, changed)
			}
		})
	}
}
