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

package audit

import (
	"testing"
)

func TestExtractEventStreamCompletionAndFcName(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		wantCompletion string
		wantFcName     string
	}{
		{
			name: "chat completions streaming",
			body: `data: {"id":"chatcmpl-1","choices":[{"delta":{"role":"assistant","content":"Hello"},"index":0}]}
data: {"id":"chatcmpl-1","choices":[{"delta":{"content":", world!"},"index":0}]}
data: [DONE]
`,
			wantCompletion: "Hello, world!",
		},
		{
			name: "responses api text.delta",
			body: `event: response.created
data: {"type":"response.created","response":{"id":"resp_1","status":"in_progress","output":[]}}

event: response.text.delta
data: {"type":"response.text.delta","delta":"Hello"}

event: response.text.delta
data: {"type":"response.text.delta","delta":", world!"}

event: response.done
data: {"type":"response.done","response":{"id":"resp_1","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Hello, world!"}]}]}}
`,
			wantCompletion: "Hello, world!",
		},
		{
			name: "responses api done only (no deltas stored)",
			body: `event: response.created
data: {"type":"response.created","response":{"id":"resp_1","status":"in_progress","output":[]}}

event: response.done
data: {"type":"response.done","response":{"id":"resp_1","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Final answer here."}]}]}}
`,
			wantCompletion: "Final answer here.",
		},
		{
			name: "responses api output_text.delta (doubao/other providers)",
			body: `event: response.output_text.delta
data: {"type":"response.output_text.delta","delta":"Hello"}

event: response.output_text.delta
data: {"type":"response.output_text.delta","delta":", world!"}
`,
			wantCompletion: "Hello, world!",
		},
		{
			name: "responses api response.completed alias",
			body: `event: response.completed
data: {"type":"response.completed","response":{"id":"resp_1","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Final answer."}]}]}}
`,
			wantCompletion: "Final answer.",
		},
		{
			name: "responses api content_part.delta (delta is object)",
			body: `event: response.content_part.delta
data: {"type":"response.content_part.delta","delta":{"type":"text","text":"Hello"}}

event: response.content_part.delta
data: {"type":"response.content_part.delta","delta":{"type":"text","text":", world!"}}
`,
			wantCompletion: "Hello, world!",
		},
		{
			name: "responses api function call arguments",
			body: `event: response.function_call_arguments.delta
data: {"type":"response.function_call_arguments.delta","name":"my_func","delta":"{\"key\":"}

event: response.function_call_arguments.delta
data: {"type":"response.function_call_arguments.delta","delta":"\"value\"}"}
`,
			wantCompletion: `{"key":"value"}`,
			wantFcName:     "my_func",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completion, fcName := ExtractEventStreamCompletionAndFcName(tt.body)
			if completion != tt.wantCompletion {
				t.Errorf("completion: got %q, want %q", completion, tt.wantCompletion)
			}
			if fcName != tt.wantFcName {
				t.Errorf("fcName: got %q, want %q", fcName, tt.wantFcName)
			}
		})
	}
}

func TestExtractResponsesAPIJsonCompletion(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		wantCompletion string
		wantFcName     string
	}{
		{
			name:           "text output",
			body:           `{"id":"resp_1","object":"response","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Hello, world!"}]}]}`,
			wantCompletion: "Hello, world!",
		},
		{
			name:           "function call output",
			body:           `{"id":"resp_1","object":"response","output":[{"type":"function_call","name":"my_func","arguments":"{\"key\":\"value\"}"}]}`,
			wantCompletion: `{"key":"value"}`,
			wantFcName:     "my_func",
		},
		{
			name:           "no output field (chat completions json)",
			body:           `{"choices":[{"message":{"content":"hello"}}]}`,
			wantCompletion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completion, fcName := ExtractResponsesAPIJsonCompletion(tt.body)
			if completion != tt.wantCompletion {
				t.Errorf("completion: got %q, want %q", completion, tt.wantCompletion)
			}
			if fcName != tt.wantFcName {
				t.Errorf("fcName: got %q, want %q", fcName, tt.wantFcName)
			}
		})
	}
}
