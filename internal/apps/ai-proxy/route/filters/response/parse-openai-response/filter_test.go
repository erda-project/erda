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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	audittypes "github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
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
			name: "responses api done fallback includes refusal content",
			body: `event: response.done
data: {"type":"response.done","response":{"id":"resp_1","output":[{"type":"message","role":"assistant","content":[{"type":"refusal","refusal":"Cannot comply."}]}]}}
`,
			wantCompletion: "Cannot comply.",
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
			name: "responses api content_part.delta ignores non text delta",
			body: `event: response.content_part.delta
data: {"type":"response.content_part.delta","delta":{"type":"input_text","text":"ignore"}}

event: response.content_part.delta
data: {"type":"response.content_part.delta","delta":{"type":"output_text","text":"keep"}}
`,
			wantCompletion: "keep",
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

func TestOnBodyChunkRecordsStreamingCompletionBeforeComplete(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutIsStream(ctx, true)
	ctxhelper.PutAuditID(ctx, "audit-1")
	ctxhelper.PutAuditSink(ctx, audittypes.New("audit-1", logrusx.New()))

	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/chat/completions", nil).WithContext(ctx)
	resp := &http.Response{Request: req}
	filter := &Filter{}

	chunk1 := []byte("data: {\"id\":\"chatcmpl-1\",\"choices\":[{\"delta\":{\"content\":\"Hello\"},\"index\":0}]}\n")
	if out, err := filter.OnBodyChunk(resp, chunk1, 0); err != nil {
		t.Fatalf("OnBodyChunk chunk1 error: %v", err)
	} else if string(out) != string(chunk1) {
		t.Fatalf("OnBodyChunk chunk1 output: got %q, want %q", out, chunk1)
	}

	if filter.completion != "Hello" {
		t.Fatalf("completion after chunk1: got %q, want %q", filter.completion, "Hello")
	}
	if filter.lastRecordedCompletionLen != len("Hello") {
		t.Fatalf("lastRecordedCompletionLen after chunk1: got %d, want %d", filter.lastRecordedCompletionLen, len("Hello"))
	}

	sink, ok := ctxhelper.GetAuditSink(ctx)
	if !ok {
		t.Fatal("expected audit sink in context")
	}
	if got := sink.Snapshot()["completion"]; got != "Hello" {
		t.Fatalf("audit completion after chunk1: got %#v, want %q", got, "Hello")
	}

	chunk2 := []byte("data: {\"id\":\"chatcmpl-1\",\"choices\":[{\"delta\":{\"content\":\", world!\"},\"index\":0}]}\n")
	if out, err := filter.OnBodyChunk(resp, chunk2, 1); err != nil {
		t.Fatalf("OnBodyChunk chunk2 error: %v", err)
	} else if string(out) != string(chunk2) {
		t.Fatalf("OnBodyChunk chunk2 output: got %q, want %q", out, chunk2)
	}

	if filter.completion != "Hello, world!" {
		t.Fatalf("completion after chunk2: got %q, want %q", filter.completion, "Hello, world!")
	}
	if filter.lastRecordedCompletionLen != len("Hello, world!") {
		t.Fatalf("lastRecordedCompletionLen after chunk2: got %d, want %d", filter.lastRecordedCompletionLen, len("Hello, world!"))
	}
	if got := sink.Snapshot()["completion"]; got != "Hello, world!" {
		t.Fatalf("audit completion after chunk2: got %#v, want %q", got, "Hello, world!")
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
			name:           "refusal output",
			body:           `{"id":"resp_1","object":"response","output":[{"type":"message","role":"assistant","content":[{"type":"refusal","refusal":"Cannot comply."}]}]}`,
			wantCompletion: "Cannot comply.",
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
