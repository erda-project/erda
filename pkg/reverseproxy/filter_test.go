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

package reverseproxy_test

import (
	"bytes"
	"net/http"
	"sync"
	"testing"

	"github.com/erda-project/erda/pkg/reverseproxy"
)

func TestInfor_Method(t *testing.T) {
	mockedBody := bytes.NewBufferString("mocked-body")
	request, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api", mockedBody)
	if err != nil {
		t.Fatal(err)
	}
	request.RemoteAddr = "127.0.0.1:3344"
	request.Host = "localhost:8080"
	request.ContentLength = int64(mockedBody.Len())
	response := &http.Response{Request: request}
	response.ContentLength = 0
	response.StatusCode = http.StatusOK
	response.Status = http.StatusText(response.StatusCode)
	ctx := reverseproxy.NewContext(make(map[any]any))
	infors := []reverseproxy.HttpInfor{
		reverseproxy.NewInfor(ctx, *request),
		reverseproxy.NewInfor(ctx, request),
		reverseproxy.NewInfor(ctx, *response),
		reverseproxy.NewInfor(ctx, response),
	}
	for i, infor := range infors {
		if method := infor.Method(); method != request.Method {
			t.Errorf("%d method error", i)
		}
		if remoteAddr := infor.RemoteAddr(); remoteAddr != request.RemoteAddr {
			t.Errorf("%d remoteAddr error", i)
		}
		if host := infor.Host(); host != request.Host {
			t.Errorf("%d host error", i)
		}
		if path := infor.URL().Path; path != "/api" {
			t.Errorf("%d path error", i)
		}
		if i <= 1 {
			if contentLength := infor.ContentLength(); contentLength != request.ContentLength {
				t.Errorf("%d contentLength error", i)
			}
			if status, code := infor.Status(), infor.StatusCode(); status != "" || code != 0 {
				t.Errorf("%d status error", i)
			}
			if body := infor.Body(); body == nil {
				t.Errorf("%d Body error", i)
			}
			if buffer := infor.BodyBuffer(); buffer == nil {
				t.Errorf("%d BodyBuffer error", i)
			}
		} else {
			if contentLength := infor.ContentLength(); contentLength != response.ContentLength {
				t.Errorf("%d contentLength error", i)
			}
			if status, code := infor.Status(), infor.StatusCode(); status != response.Status || code != response.StatusCode {
				t.Errorf("%d status error", i)
			}
			if body := infor.Body(); body != nil {
				t.Errorf("%d Body error", i)
			}
			if buffer := infor.BodyBuffer(); buffer != nil {
				t.Errorf("%d BodyBuffer error", i)
			}
		}

	}
}

func TestInfor_Header(t *testing.T) {
	request, err := http.NewRequest(http.MethodPost, "http://localhost:8080", bytes.NewBufferString("mock body"))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	infor := reverseproxy.NewInfor(reverseproxy.NewContext(map[any]any{
		reverseproxy.MutexCtxKey{}: new(sync.Mutex),
	}), request)
	header := infor.Header()
	t.Logf("Content-Type: %s", header.Get("Content-Type"))
	header.Set("Accept", "*/*")
	t.Logf("Accept: %s", infor.Header().Get("Accept"))
	header.Set("Content-Type", "application/xml")
	t.Logf("Content-Type: %s", infor.Header().Get("Content-Type"))

	u := infor.URL()
	values := u.Query()
	values.Add("api-version", "gpt-3.5")
	u.RawQuery = values.Encode()

	t.Logf("request.URL.RequestURI(): %s", request.URL.RequestURI())

	t.Logf("infor.BodyBuffer(): %v", infor.BodyBuffer())
	t.Logf("infor.Body(): %v", infor.Body())
}
