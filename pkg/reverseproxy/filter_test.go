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
	"net/http"
	"sync"
	"testing"

	"github.com/erda-project/erda/pkg/reverseproxy"
)

func TestInfor_Header(t *testing.T) {
	ctx := context.WithValue(context.Background(), reverseproxy.MutexCtxKey{}, new(sync.Mutex))
	request, err := http.NewRequest(http.MethodPost, "http://localhost:8080", bytes.NewBufferString("mock body"))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
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
}
