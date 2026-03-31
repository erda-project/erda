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
	"io"
	"testing"

	set_resp_body_chunk_splitter "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/set-resp-body-chunk-splitter"
)

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

