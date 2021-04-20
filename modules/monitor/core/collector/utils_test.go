// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package collector

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const bodyContent = `[{"stream": "stderr", "tags": {"dice_component": "xxx"}, "timestamp": 1596187771340803293, "labels": {}, "content": "info hello world", "source": "container", "offset": 10095271, "id": "ef864e243d36fe518a9716964f26b85f1942a7a108bc6f62d55eb726b7d1e5b1"}]`

func TestReadRequestBody_b64_gzip(t *testing.T) {
	req := mockRequest(strings.NewReader(bodyContent))
	b64Request(req)
	gzipRequest(req)

	r, err := ReadRequestBodyReader(req)

	assert.Nil(t, err)
	body, err := ioutil.ReadAll(r)
	assert.Nil(t, err)
	assert.Equal(t, bodyContent, string(body))
}

func TestReadRequestBody_b64(t *testing.T) {
	req := mockRequest(strings.NewReader(bodyContent))
	b64Request(req)

	r, err := ReadRequestBodyReader(req)

	assert.Nil(t, err)
	body, err := ioutil.ReadAll(r)
	assert.Nil(t, err)
	assert.Equal(t, bodyContent, string(body))
}

func TestReadRequestBody(t *testing.T) {
	req := mockRequest(strings.NewReader(bodyContent))

	r, err := ReadRequestBodyReader(req)

	assert.Nil(t, err)
	body, err := ioutil.ReadAll(r)
	assert.Nil(t, err)
	assert.Equal(t, bodyContent, string(body))
}

func mockRequest(body io.Reader) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "http://localhost.com", body)
	return req
}

func gzipRequest(req *http.Request) {
	content, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	req.Body.Close()

	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(content); err != nil {
		log.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		log.Fatal(err)
	}

	body := io.Reader(bytes.NewReader(b.Bytes()))
	rc, ok := (body).(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}

	req.Body = rc
	req.Header.Set("Content-Encoding", "gzip")
}

func b64Request(req *http.Request) {
	content, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}

	body := io.Reader(strings.NewReader(base64.StdEncoding.EncodeToString(content)))
	rc, ok := (body).(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	req.Body = rc
	req.Header.Set("Custom-Content-Encoding", "base64")
}

func Test_isJSONArray(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"", args{b: []byte(`[{"a":1}]`)}, true},
		{"", args{b: []byte(`{"a":1}`)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isJSONArray(tt.args.b); got != tt.want {
				t.Errorf("isJSONArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
