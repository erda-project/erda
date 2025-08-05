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

package context

import (
	"encoding/json"
	"testing"
)

func TestFindUserPrompts(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "string input",
			input:    `"hello"`,
			expected: []string{"hello"},
		},
		{
			name: "array of user messages",
			input: `[
				{
					"role": "user",
					"content": "hello"
				},
				{
					"role": "assistant",
					"content": "hi"
				},
				{
					"role": "user",
					"content": "how are you"
				}
			]`,
			expected: []string{"hello", "how are you"},
		},
		{
			name: "array with mixed content types",
			input: `[
				{
					"role": "user",
					"content": [
						{
							"type": "input_text",
							"text": "text"
						},
						{
							"type": "input_image",
 							"image_url": "https://www.erda.cloud/_next/image?url=%2Fimages%2Flogo-new.png&w=256&q=75"
						},
						{
							"type": "input_text",
							"text": "more text"
						}
					]
				}
			]`,
			expected: []string{"text", "more text"},
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: []string{},
		},
		{
			name:     "null input",
			input:    `null`,
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var input interface{}
			if err := json.Unmarshal([]byte(tc.input), &input); err != nil {
				t.Fatalf("failed to unmarshal input: %v", err)
			}
			result := FindUserPrompts(input)
			if len(result) != len(tc.expected) {
				t.Errorf("expected %d prompts, got %d", len(tc.expected), len(result))
				return
			}
			for i, prompt := range result {
				if prompt != tc.expected[i] {
					t.Errorf("prompt %d: expected %q, got %q", i, tc.expected[i], prompt)
				}
			}
		})
	}
}
