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

package mcp

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestProcessAnyOf(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "basic anyOf replace",
			input: `{
				"type": "object",
				"anyOf": [
					{"type": "string"},
					{"type": "number"}
				]
			}`,
			expected: `{
				"type": "string"
			}`,
		},
		{
			name: "nested anyOf in properties",
			input: `{
				"type": "object",
				"properties": {
					"field": {
						"anyOf": [
							{"type": "boolean"},
							{"type": "null"}
						]
					}
				}
			}`,
			expected: `{
				"type": "object",
				"properties": {
					"field": {
						"type": "boolean"
					}
				}
			}`,
		},
		{
			name: "array with anyOf",
			input: `[
				{"anyOf": [{"format": "date-time"}, {"format": "date"}]},
				{"type": "integer"}
			]`,
			expected: `[
				{"format": "date-time"},
				{"type": "integer"}
			]`,
		},
		{
			name:     "no anyOf should remain unchanged",
			input:    `{"type": "number"}`,
			expected: `{"type": "number"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var obj interface{}
			if err := json.Unmarshal([]byte(tt.input), &obj); err != nil {
				t.Fatalf("failed to unmarshal input: %v", err)
			}

			ProcessAnyOf(obj)

			var expected interface{}
			if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
				t.Fatalf("failed to unmarshal expected: %v", err)
			}

			if !reflect.DeepEqual(obj, expected) {
				t.Errorf("result mismatch\n got: %+v\nwant: %+v", obj, expected)
			}
		})
	}
}
