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

package decompress

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/http/httputil"
)

func TryDecompressBody(header http.Header, r *bytes.Buffer) []byte {
	result, err := DecompressBody(header, r)
	if err != nil {
		logrus.Warnf("decompress body failed: %v", err)
		return nil // avoid to return the uncompressed body, which will let db failed to save the data
	}
	return result
}

// DecompressBody decompresses the body according to the content encoding.
// If no content encoding is specified, the body is returned as is.
func DecompressBody(header http.Header, r *bytes.Buffer) ([]byte, error) {
	if r == nil || r.Bytes() == nil {
		return nil, nil
	}
	encodings := header.Get(httputil.HeaderKeyContentEncoding)
	for _, encoding := range strings.Split(encodings, ",") {
		encoding = strings.TrimSpace(encoding)
		if len(encoding) == 0 {
			continue
		}
		var rr io.Reader
		var err error
		switch encoding {
		case "gzip":
			rr, err = gzip.NewReader(r)
		case "deflate":
			rr, err = zlib.NewReader(r)
		default:
			return nil, fmt.Errorf("unsupported content encoding: %s", encoding)
		}
		if err != nil {
			return nil, err
		}
		return io.ReadAll(rr)
	}
	// not compressed, return as is
	return r.Bytes(), nil
}
