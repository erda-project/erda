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

package body_util

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

// const MaxSample = 4 * 1024 // Sample at most 4KB
const MaxSample = -1

// SmartCloneBody only caches sampled portion, main chain continues with remaining stream
// Note: body must be *io.ReadCloser, as it needs to be replaced for main chain
func SmartCloneBody(body *io.ReadCloser, maxSample int64) (snapshot ReadCloser, err error) {
	var buf []byte
	var n int
	if maxSample >= 0 {
		buf = make([]byte, maxSample)
		n, err = io.ReadFull(*body, buf)
		if err != nil && err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, err
		}
	} else {
		b, err := io.ReadAll(*body)
		if err != nil {
			return nil, err
		}
		buf, n = b, len(b)
	}
	// Observer snapshot, only return sampled portion
	snapshot = &readCloser{
		ReadCloser: io.NopCloser(bytes.NewReader(buf[:n])),
		size:       int64(n),
	}

	// Use MultiReader to construct complete body
	restReader := *body
	*body = io.NopCloser(io.MultiReader(bytes.NewReader(buf[:n]), restReader))
	return snapshot, nil
}

// SafeCloneRequest for observer filter, combined with smartCloneBody
func SafeCloneRequest(orig *http.Request, maxSample int64) (snapshot http.Request, err error) {
	// 1. Optimize with smartCloneBody
	snapBody, err := SmartCloneBody(&orig.Body, maxSample)
	if err != nil {
		return http.Request{}, err
	}
	// 2. Construct observer snapshot
	snapshot = *orig.Clone(orig.Context())
	snapshot.Body = snapBody
	return snapshot, nil
}
