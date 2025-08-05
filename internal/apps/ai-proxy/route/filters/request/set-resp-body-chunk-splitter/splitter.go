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

// Package set_resp_body_chunk_splitter provides unified ChunkSplitter interface and common implementations.
// Any protocol that needs custom chunking rules can implement ChunkSplitter and inject via
//
//	ctxhelper.PutRespBodyChunkSplitter(ctx, mySplitter)
//
// into the request's Context.
package set_resp_body_chunk_splitter

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

// ChunkSplitter splits continuous io.Reader byte stream into "logical chunks".
// NextChunk rules:
//  1. Normal return (chunk, nil) indicates reading a complete chunk;
//  2. Return (chunk, io.EOF) represents stream end, but still carries the last chunk of data;
//  3. chunk == nil && err == io.EOF represents complete end;
//  4. Other err will be returned as 502 by framework to client.
type ChunkSplitter = filter_define.RespBodyChunkSplitter

// -------- Built-in implementations -----------------------------------------------------------

// FixedSizeSplitter reads fixed Size bytes each time; suitable as fallback solution.
type FixedSizeSplitter struct{ Size int }

func (s *FixedSizeSplitter) NextChunk(r io.Reader) ([]byte, error) {
	if s.Size <= 0 {
		s.Size = 32 << 10 // Default 32 KiB
	}
	buf := make([]byte, s.Size)
	n, err := io.ReadAtLeast(r, buf, 1)
	return buf[:n], err
}

// NewLineSplitter splits chunks by single '\n' terminator; return value includes the '\n'.
type NewLineSplitter struct {
	scanner *bufio.Scanner
}

func (s *NewLineSplitter) NextChunk(r io.Reader) ([]byte, error) {
	if s.scanner == nil {
		s.scanner = bufio.NewScanner(r)
	}
	if !s.scanner.Scan() {
		if err := s.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	line := append(s.scanner.Bytes(), '\n')
	return line, nil
}

// SSESplitter for text/event-stream (Server-Sent Events),
// uses double newline "\n\n" as delimiter, return value preserves the delimiter itself.
type SSESplitter struct {
	buf bytes.Buffer
}

func (s *SSESplitter) NextChunk(r io.Reader) ([]byte, error) {
	for {
		if idx := bytes.Index(s.buf.Bytes(), []byte("\n\n")); idx >= 0 {
			chunk := make([]byte, idx+2) // Preserve \n\n
			_, _ = s.buf.Read(chunk)
			return chunk, nil
		}
		tmp := make([]byte, 4096)
		n, err := r.Read(tmp)
		if n > 0 {
			s.buf.Write(tmp[:n])
		}
		if err != nil {
			if err == io.EOF && s.buf.Len() > 0 {
				out := s.buf.Bytes()
				s.buf.Reset()
				return out, io.EOF
			}
			return nil, err
		}
	}
}

// WholeStreamSplitter reads the entire stream at once; suitable for non-streaming scenarios.
type WholeStreamSplitter struct{}

func (WholeStreamSplitter) NextChunk(r io.Reader) ([]byte, error) {
	b, err := io.ReadAll(r)
	if err == nil {
		if len(b) == 0 {
			return nil, io.EOF // Empty stream
		}
		err = io.EOF // Last chunk + EOF
	}
	return b, err
}

type BedrockStreamSplitter struct {
	br *bufio.Reader
}

func (s *BedrockStreamSplitter) NextChunk(r io.Reader) ([]byte, error) {
	if s.br == nil {
		s.br = bufio.NewReader(r)
	}
	// 1) Read first 8 bytes
	header, err := s.br.Peek(8)
	if err != nil {
		return nil, err
	}
	total := binary.BigEndian.Uint32(header[0:4])
	// 2) Read entire frame
	frame := make([]byte, total)
	if _, err := io.ReadFull(s.br, frame); err != nil {
		return nil, err
	}
	return frame, nil
}
