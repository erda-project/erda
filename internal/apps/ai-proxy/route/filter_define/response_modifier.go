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

package filter_define

import (
	"io"
	"net/http"
)

// RespBodyChunkSplitter splits continuous io.Reader byte stream into "logical chunks".
// NextChunk rules:
//  1. Normal return (chunk, nil) indicates reading a complete chunk;
//  2. Return (chunk, io.EOF) represents stream end, but still carries the last chunk of data;
//  3. chunk == nil && err == io.EOF represents complete end;
//  4. Other err will be returned as 502 by framework to client.
type RespBodyChunkSplitter interface {
	NextChunk(r io.Reader) ([]byte, error)
}

// SplitterKey used to attach specific implementation to req.Context(),
// myResponseModify extracts from it and drives chunking.
type SplitterKey struct{}

// -----------------------------------------------------------------------------

// ProxyResponseModifier performs three-phase processing on backend response: Header → Chunk → End.
// No distinction between streaming / non-streaming:
//   - Non-streaming scenarios only receive OnBodyChunk once, then OnComplete.
//   - Streaming scenarios trigger OnBodyChunk once for each logical chunk.
//
// Design points:
//  1. **ctx**: Context throughout the same request, can access timeout, logs, TraceID, etc.
//  2. **OnBodyChunk** return values:
//     - out  → bytes written to downstream (nil means swallowed);
//     - err  → terminates entire chain; framework will CloseWithError and return 502.
type ProxyResponseModifier interface {
	// OnHeaders called once after receiving backend response headers, before starting to read body.
	// Can modify headers (add/delete/modify) and do initialization, such as gzip.Reader creation.
	OnHeaders(resp *http.Response) error

	// OnBodyChunk called when each logical chunk is ready.
	// Needs to be synchronous: if processing takes long time, should split into goroutines yourself.
	OnBodyChunk(resp *http.Response, chunk []byte, index int64) (out []byte, err error)

	// OnComplete called on EOF or when framework encounters read error (called even if previous two steps error,
	// ensuring cleanup). Can write Trailer, output summary statistics, close internal resources, etc.
	// If return value is non-nil, also triggers 502.
	OnComplete(resp *http.Response) (out []byte, err error)
}

type ProxyResponseModifierEnabler interface {
	Enable(resp *http.Response) bool
}

type ProxyResponseModifierBodyPeeker interface {
	// OnPeekChunkBeforeHeaders inspects the first chunk of the response body before headers are sent.
	OnPeekChunkBeforeHeaders(resp *http.Response, peekedBody []byte) error
}

// PassThroughResponseModifier passes response as-is to next Filter.
// After embedding, just override specific methods as needed.
type PassThroughResponseModifier struct{}

func (m *PassThroughResponseModifier) Enable(resp *http.Response) bool {
	return true
}
func (m *PassThroughResponseModifier) OnHeaders(resp *http.Response) error {
	return nil
}
func (m *PassThroughResponseModifier) OnBodyChunk(resp *http.Response, chunk []byte, index int64) (out []byte, err error) {
	return chunk, nil
}
func (m *PassThroughResponseModifier) OnComplete(resp *http.Response) (out []byte, err error) {
	return nil, nil
}
func (m *PassThroughResponseModifier) OnPeekChunkBeforeHeaders(resp *http.Response, peekedBody []byte) error {
	return nil
}
