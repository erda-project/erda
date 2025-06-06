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
	"fmt"
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
)

type zstdReadCloser struct {
	*zstd.Decoder
}

func (z *zstdReadCloser) Close() error {
	z.Decoder.Close()
	return nil
}

type nopWriteCloser struct {
	io.Writer
}

func (nwc nopWriteCloser) Close() error {
	return nil
}

func NewBodyDecompressor(header http.Header, body io.Reader) (io.ReadCloser, error) {
	ce := header.Get("Content-Encoding")
	switch ce {
	case "gzip":
		return gzip.NewReader(body)
	case "deflate":
		r := flate.NewReader(body)
		return r, nil
	case "br":
		r := brotli.NewReader(body)
		return io.NopCloser(r), nil
	case "zstd":
		r, err := zstd.NewReader(body)
		rr := &zstdReadCloser{Decoder: r}
		return rr, err
	case "identity", "":
		if _, ok := body.(io.Closer); ok {
			return body.(io.ReadCloser), nil
		}
		return io.NopCloser(body), nil
	default:
		return nil, fmt.Errorf("unsupported Content-Encoding when decompress body: %s", ce)

	}
}

func NewBodyCompressor(header http.Header, body io.Writer) (io.WriteCloser, error) {
	ce := header.Get("Content-Encoding")
	switch ce {
	case "gzip":
		return gzip.NewWriterLevel(body, gzip.DefaultCompression)
	case "deflate":
		return flate.NewWriter(body, flate.DefaultCompression)
	case "br":
		return brotli.NewWriterLevel(body, brotli.DefaultCompression), nil
	case "zstd":
		return zstd.NewWriter(body, zstd.WithEncoderLevel(zstd.SpeedDefault))
	case "identity", "":
		if _, ok := body.(io.Closer); ok {
			return body.(io.WriteCloser), nil
		}
		return nopWriteCloser{Writer: body}, nil
	default:
		return nil, fmt.Errorf("unsupported Content-Encoding when compress body: %s", ce)
	}
}
