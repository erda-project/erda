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

package common

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang/snappy"
)

func NormalizeKey(key string) string {
	key = strings.ReplaceAll(key, ".", "_")
	key = strings.ReplaceAll(key, "/", "_")
	return key
}

func IsJSONArray(b []byte) bool {
	x := bytes.TrimLeft(b, " \t\r\n")
	return len(x) > 0 && x[0] == '['
}

// read request's body based on Content-Encoding Header
func ReadBody(req *http.Request) ([]byte, error) {
	encoding := req.Header.Get("Content-Encoding")
	defer req.Body.Close()

	switch encoding {
	case "gzip":
		r, err := gzip.NewReader(req.Body)
		if err != nil {
			return nil, fmt.Errorf("gzip.NewReader err: %w", err)
		}

		bytes, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	case "snappy":
		bytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		// snappy block format is only supported by decode/encode not snappy reader/writer
		bytes, err = snappy.Decode(nil, bytes)
		if err != nil {
			return nil, fmt.Errorf("snappy.Decode err: %w", err)
		}
		return bytes, nil
	default:
		bytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	}
}
