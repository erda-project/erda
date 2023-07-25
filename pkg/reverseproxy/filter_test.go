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
	"context"
	"io"
	"net/http"
	"sync"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
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
	ctx := context.Background()
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
	ctx := context.WithValue(context.Background(), reverseproxy.MutexCtxKey{}, new(sync.Mutex))
	infor := reverseproxy.NewInfor(ctx, request)
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

func TestInfor_SetBody(t *testing.T) {
	t.Run("mocked request body", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodPost, "http://localhost:8080", bytes.NewBufferString("mock body"))
		if err != nil {
			t.Fatal(err)
		}
		infor := reverseproxy.NewInfor(context.Background(), request)
		infor.SetBody(io.NopCloser(bytes.NewBufferString("new body")))
		data, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer request.Body.Close()
		if string(data) != "new body" {
			t.Fatal("body data err")
		}
	})
	t.Run("nil request body", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodPost, "http://localhost:8080", nil)
		if err != nil {
			t.Fatal(err)
		}
		infor := reverseproxy.NewInfor(context.Background(), request)
		infor.SetBody(io.NopCloser(bytes.NewBufferString("new body")))
		data, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer request.Body.Close()
		if string(data) != "new body" {
			t.Fatal("body data err")
		}
	})
	t.Run("mocked response body", func(t *testing.T) {
		response := &http.Response{Body: io.NopCloser(bytes.NewBufferString("mocked body"))}
		infor := reverseproxy.NewInfor(context.Background(), response)
		infor.SetBody(io.NopCloser(bytes.NewBufferString("new body")))
		data, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if string(data) != "new body" {
			t.Fatal("body data err")
		}
	})
	t.Run("nil response body", func(t *testing.T) {
		response := new(http.Response)
		infor := reverseproxy.NewInfor(context.Background(), response)
		infor.SetBody(io.NopCloser(bytes.NewBufferString("new body")))
		data, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if string(data) != "new body" {
			t.Fatal("body data err")
		}
	})
}

func TestFilterConfig(t *testing.T) {
	var y = `
name: session-context
config:
  "on":
    - key: X-Ai-Proxy-Source
      operator: =
      value: erda.cloud
`
	var fc reverseproxy.FilterConfig
	if err := yaml.Unmarshal([]byte(y), &fc); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", strutil.TryGetYamlStr(fc))
}
