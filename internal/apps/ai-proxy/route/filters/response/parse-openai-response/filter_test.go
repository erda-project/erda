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

func TestExtractEventStreamCompletionAndFcName_ChatCompletions(t *testing.T) {
	body := `data: {"id":"chatcmpl-1","choices":[{"delta":{"role":"assistant","content":"Hello"},"index":0}]}
data: {"id":"chatcmpl-1","choices":[{"delta":{"content":", world!"},"index":0}]}
data: [DONE]
`
	completion, _ := ExtractEventStreamCompletionAndFcName(body)
	if completion != "Hello, world!" {
		t.Errorf("expected 'Hello, world!' got %q", completion)
	}
}

func TestExtractEventStreamCompletionAndFcName_ResponsesAPIDeltas(t *testing.T) {
	body := `event: response.created
data: {"type":"response.created","response":{"id":"resp_1","status":"in_progress","output":[]}}

event: response.text.delta
data: {"type":"response.text.delta","delta":"Hello"}

event: response.text.delta
data: {"type":"response.text.delta","delta":", world!"}

event: response.done
data: {"type":"response.done","response":{"id":"resp_1","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Hello, world!"}]}]}}
`
	completion, _ := ExtractEventStreamCompletionAndFcName(body)
	if completion != "Hello, world!" {
		t.Errorf("expected 'Hello, world!' got %q", completion)
	}
}

func TestExtractEventStreamCompletionAndFcName_ResponsesAPIDoneOnly(t *testing.T) {
	// simulate case where only response.created and response.done are stored (no deltas)
	body := `event: response.created
data: {"type":"response.created","response":{"id":"resp_1","status":"in_progress","output":[]}}

event: response.done
data: {"type":"response.done","response":{"id":"resp_1","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Final answer here."}]}]}}
`
	completion, _ := ExtractEventStreamCompletionAndFcName(body)
	if completion != "Final answer here." {
		t.Errorf("expected 'Final answer here.' got %q", completion)
	}
}

func TestExtractEventStreamCompletionAndFcName_ResponsesAPIOutputTextDelta(t *testing.T) {
	// response.output_text.delta (used by Doubao/other providers)
	body := `event: response.output_text.delta
data: {"type":"response.output_text.delta","delta":"Hello"}

event: response.output_text.delta
data: {"type":"response.output_text.delta","delta":", world!"}
`
	completion, _ := ExtractEventStreamCompletionAndFcName(body)
	if completion != "Hello, world!" {
		t.Errorf("expected 'Hello, world!' got %q", completion)
	}
}

func TestExtractEventStreamCompletionAndFcName_ResponsesAPICompleted(t *testing.T) {
	// response.completed is an alias for response.done used by some providers
	body := `event: response.completed
data: {"type":"response.completed","response":{"id":"resp_1","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Final answer."}]}]}}
`
	completion, _ := ExtractEventStreamCompletionAndFcName(body)
	if completion != "Final answer." {
		t.Errorf("expected 'Final answer.' got %q", completion)
	}
}

func TestExtractEventStreamCompletionAndFcName_ResponsesAPIContentPartDelta(t *testing.T) {
	// response.content_part.delta has delta as an object, not a plain string
	body := `event: response.content_part.delta
data: {"type":"response.content_part.delta","delta":{"type":"text","text":"Hello"}}

event: response.content_part.delta
data: {"type":"response.content_part.delta","delta":{"type":"text","text":", world!"}}
`
	completion, _ := ExtractEventStreamCompletionAndFcName(body)
	if completion != "Hello, world!" {
		t.Errorf("expected 'Hello, world!' got %q", completion)
	}
}

func TestExtractEventStreamCompletionAndFcName_ResponsesAPIFunctionCall(t *testing.T) {
	body := `event: response.function_call_arguments.delta
data: {"type":"response.function_call_arguments.delta","name":"my_func","delta":"{\"key\":"}

event: response.function_call_arguments.delta
data: {"type":"response.function_call_arguments.delta","delta":"\"value\"}"}
`
	completion, fcName := ExtractEventStreamCompletionAndFcName(body)
	if fcName != "my_func" {
		t.Errorf("expected fcName 'my_func' got %q", fcName)
	}
	if completion != `{"key":"value"}` {
		t.Errorf("expected completion %q got %q", `{"key":"value"}`, completion)
	}
}
