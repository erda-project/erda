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

package reverse_proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	set_resp_body_chunk_splitter "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/set-resp-body-chunk-splitter"
)

func TestTruncateBodyForAudit(t *testing.T) {
	body := []byte("abcdefghijklmnopqrstuvwxyz")
	got := truncateBodyForAudit(body, 5, 3)
	want := "abcde...[omitted 18 bytes]...xyz"
	if string(got) != want {
		t.Fatalf("unexpected truncation result: %q", string(got))
	}
}

func TestSSESplitter_PeekAndContinue(t *testing.T) {
	sseData := "data: {\"id\":\"1\"}\n\ndata: {\"id\":\"2\"}\n\ndata: [DONE]\n\n"
	reader := io.NopCloser(bytes.NewReader([]byte(sseData)))
	splitter := &set_resp_body_chunk_splitter.SSESplitter{}

	chunk1, err1 := splitter.NextChunk(reader)
	if err1 != nil {
		t.Fatalf("unexpected error on first chunk: %v", err1)
	}
	if string(chunk1) != "data: {\"id\":\"1\"}\n\n" {
		t.Fatalf("unexpected first chunk: %q", string(chunk1))
	}

	chunk2, err2 := splitter.NextChunk(reader)
	if err2 != nil {
		t.Fatalf("unexpected error on second chunk: %v", err2)
	}
	if string(chunk2) != "data: {\"id\":\"2\"}\n\n" {
		t.Fatalf("unexpected second chunk: %q", string(chunk2))
	}

	chunk3, err3 := splitter.NextChunk(reader)
	if err3 != nil && err3 != io.EOF {
		t.Fatalf("unexpected error on third chunk: %v", err3)
	}
	if string(chunk3) != "data: [DONE]\n\n" {
		t.Fatalf("unexpected third chunk: %q", string(chunk3))
	}
}

func TestWholeStreamSplitter_PeekReadsAll(t *testing.T) {
	jsonData := `{"choices":[{"message":{"content":"hello"}}]}`
	reader := io.NopCloser(bytes.NewReader([]byte(jsonData)))
	splitter := &set_resp_body_chunk_splitter.WholeStreamSplitter{}

	chunk, err := splitter.NextChunk(reader)
	if err != io.EOF {
		t.Fatalf("expected io.EOF, got: %v", err)
	}
	if string(chunk) != jsonData {
		t.Fatalf("unexpected chunk: %q", string(chunk))
	}

	chunk2, err2 := splitter.NextChunk(reader)
	if err2 != io.EOF {
		t.Fatalf("expected io.EOF on second read, got: %v", err2)
	}
	if len(chunk2) != 0 {
		t.Fatalf("expected empty chunk, got: %q", string(chunk2))
	}
}

// makeSSEBody joins SSE events with the standard \n\n delimiter.
// The body is pure SSE — no HTTP headers, matching what asyncHandleRespBody receives.
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
		name      string
		body      string
		headLimit int
		tailLimit int
		contains  []string
		notContains []string
	}{
		{
			name:        "non-SSE falls back to truncation",
			body:        `{"choices":[{"message":{"content":"hello"}}]}`,
			headLimit:   5,
			tailLimit:   3,
			contains:    []string{"omitted"},
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
			headLimit:   10,
			tailLimit:   5,
			contains:    []string{"omitted"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(optimizeBodyForAudit([]byte(tt.body), tt.headLimit, tt.tailLimit))
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
