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

package httputil_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/erda-project/erda/pkg/http/httputil"
)

func TestNopCloseReadBody(t *testing.T) {
	request, err := http.NewRequest(http.MethodPost, "https://localhost:8080", bytes.NewBuffer([]byte("this is the body data")))
	if err != nil {
		t.Fatal(err)
	}
	data, err := httputil.NopCloseReadBody(request)
	if err != nil {
		t.Fatal(err)
	}

	raw, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("clone data: %s", data.String())
	t.Logf("raw data  : %s", string(raw))
}

func TestNopCloseReadBody2(t *testing.T) {
	var buf = new(bytes.Buffer)
	request, err := http.NewRequest(http.MethodPost, "https://localhost:8080", buf)
	if err != nil {
		t.Fatal(err)
	}
	data, err := httputil.NopCloseReadBody(request)
	if err != nil {
		t.Fatal(err)
	}

	raw, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("clone data: %s", data.String())
	t.Logf("raw data  : %s", string(raw))
}
