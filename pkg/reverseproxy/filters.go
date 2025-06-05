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

package reverseproxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
)

var (
	_ ResponseFilter = (*DefaultResponseFilter)(nil)
	_ ResponseFilter = (*responseBodyWriter)(nil)
)

func NewDefaultResponseFilter() *DefaultResponseFilter {
	return &DefaultResponseFilter{Buffer: bytes.NewBuffer(nil)}
}

// DefaultResponseFilter 的 OnResponseChunk 和 OnResponseEOF 将每一片 chunk 记录在自身的 buffer 中, 没有做其他任何事情.
type DefaultResponseFilter struct {
	*bytes.Buffer
}

// OnResponseChunk 将传入的 chunk 记录在自身的 buffer 中, 同时通过写入 w io.Writer 的方式传递给下一个 ResponseFilter, 没有做其他任何事情.
func (d *DefaultResponseFilter) OnResponseChunk(ctx context.Context, _ HttpInfor, w Writer, chunk []byte) (signal Signal, err error) {
	err = d.multiWrite(w, chunk)
	if err != nil {
		if l, ok := ctx.Value(LoggerCtxKey{}).(interface{ Errorf(string, ...interface{}) }); ok {
			l.Errorf("failed to multiWrite OnResponseChunk, err: %v", err)
		}
	}
	return map[bool]Signal{true: Continue, false: Intercept}[err == nil], err
}
func (d *DefaultResponseFilter) OnResponseChunkImmutable(ctx context.Context, _ HttpInfor, copiedChunk []byte) (signal Signal, err error) {
	return Continue, nil
}

// OnResponseEOF 将传入的 chunk 记录在自身的 buffer 中, 同时通过写入 w io.Writer 的方式传递给下一个 ResponseFilter, 没有做其他任何事情.
func (d *DefaultResponseFilter) OnResponseEOF(ctx context.Context, _ HttpInfor, w Writer, chunk []byte) (err error) {
	err = d.multiWrite(w, chunk)
	if err != nil {
		if l, ok := ctx.Value(LoggerCtxKey{}).(interface{ Errorf(string, ...interface{}) }); ok {
			l.Errorf("failed to multiWrite OnResponseEOF, err: %v", err)
		}
	}
	return err
}
func (d *DefaultResponseFilter) OnResponseEOFImmutable(ctx context.Context, _ HttpInfor, copiedChunk []byte) (err error) {
	return nil
}

func (d *DefaultResponseFilter) multiWrite(w io.Writer, in []byte) error {
	_, err := io.MultiWriter(d, w).Write(in)
	return err
}

type responseBodyWriter struct {
	written int64
	dst     io.Writer
}

func (r *responseBodyWriter) OnResponseChunk(ctx context.Context, infor HttpInfor, writer Writer, chunk []byte) (signal Signal, err error) {
	return Intercept, r.compressWrite(infor.Header(), chunk)
}
func (r *responseBodyWriter) OnResponseChunkImmutable(ctx context.Context, infor HttpInfor, copiedChunk []byte) (signal Signal, err error) {
	return Continue, nil
}

// OnResponseEOF responseBodyWriter 是一个特殊 ResponseFilter, 它不将数据写入传入的 io.Writer (即传递到下一个 filter),
// 因为它已经没有下一个 filter 了.
// 它直接将数据写入 r.dst, 这个 r.dst 即最终的 response body.
func (r *responseBodyWriter) OnResponseEOF(ctx context.Context, infor HttpInfor, writer Writer, chunk []byte) error {
	return r.compressWrite(infor.Header(), chunk)
}
func (r *responseBodyWriter) OnResponseEOFImmutable(ctx context.Context, infor HttpInfor, copiedChunk []byte) error {
	return nil
}

func (r *responseBodyWriter) compressWrite(header http.Header, in []byte) error {
	// compress body by header
	dstCompressor, err := CompressBody(header, r.dst)
	if err != nil {
		return fmt.Errorf("failed to compress body: %v", err)
	}
	defer dstCompressor.Close()

	_, err = dstCompressor.Write(in)
	return err
}

func (r *responseBodyWriter) write(in []byte) error {
	n, err := r.dst.Write(in)
	if n > 0 {
		r.written += int64(n)
	}
	if err != nil {
		return err
	}
	if n != len(in) {
		log.Println(io.ErrShortWrite)
		return io.ErrShortWrite
	}
	return nil
}
