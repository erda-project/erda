// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package util

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"gotest.tools/assert"
)

func TestHandleRequest(t *testing.T) {
	req := &http.Request{
		Body:   io.NopCloser(bytes.NewReader([]byte("aGVsbG8gZXJkYQ=="))),
		Header: map[string][]string{},
	}
	req.Header.Set("base64-encoded-request-body", "true")
	HandleRequest(req)
	b, _ := ioutil.ReadAll(req.Body)
	assert.Equal(t, "hello erda", string(b))
}
