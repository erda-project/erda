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
	"bufio"
	"io"
	"net"
	"net/http"
)

// TrackedResponseWriter wraps http.ResponseWriter and tracks whether headers
// have been written. It's used to ensure retries are only attempted before
// any bytes are sent to the client.
type TrackedResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
	statusCode  int
}

func NewTrackedResponseWriter(w http.ResponseWriter) *TrackedResponseWriter {
	if tw, ok := w.(*TrackedResponseWriter); ok {
		return tw
	}
	return &TrackedResponseWriter{ResponseWriter: w}
}

func (w *TrackedResponseWriter) WriteHeader(statusCode int) {
	w.wroteHeader = true
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *TrackedResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		// implicit WriteHeader(http.StatusOK)
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func (w *TrackedResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(r)
	}
	return io.Copy(w.ResponseWriter, r)
}

func (w *TrackedResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *TrackedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return h.Hijack()
}

func (w *TrackedResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (w *TrackedResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *TrackedResponseWriter) WroteHeader() bool { return w.wroteHeader }

func (w *TrackedResponseWriter) StatusCode() int { return w.statusCode }
