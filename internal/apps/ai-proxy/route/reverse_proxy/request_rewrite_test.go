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

package reverse_proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func TestHandleAIProxyRequestHeaderSetsTrustedHealthProbeForLoopback(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8081/v1/chat/completions", nil).WithContext(ctx)
	req.RemoteAddr = "127.0.0.1:43123"
	req.Header.Set(vars.XAIProxyModelHealthProbe, "true")
	out := req.Clone(ctx)
	pr := &httputil.ProxyRequest{In: req, Out: out}

	handleAIProxyRequestHeader(pr)

	trusted, ok := ctxhelper.GetTrustedHealthProbe(ctx)
	if !ok || !trusted {
		t.Fatal("expected trusted health probe set in context for loopback probe request")
	}
}

func TestHandleAIProxyRequestHeaderDoesNotSetTrustedHealthProbeForNonLoopback(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8081/v1/chat/completions", nil).WithContext(ctx)
	req.RemoteAddr = "10.20.30.40:43123"
	req.Header.Set(vars.XAIProxyModelHealthProbe, "true")
	out := req.Clone(ctx)
	pr := &httputil.ProxyRequest{In: req, Out: out}

	handleAIProxyRequestHeader(pr)

	if trusted, ok := ctxhelper.GetTrustedHealthProbe(ctx); ok && trusted {
		t.Fatal("did not expect trusted health probe for non-loopback request")
	}
}

func TestIsLoopbackRemoteAddr(t *testing.T) {
	cases := []struct {
		addr string
		want bool
	}{
		{addr: "127.0.0.1:1234", want: true},
		{addr: "[::1]:1234", want: true},
		{addr: "10.0.0.1:1234", want: false},
		{addr: "", want: false},
	}
	for _, tt := range cases {
		if got := isLoopbackRemoteAddr(tt.addr); got != tt.want {
			t.Fatalf("isLoopbackRemoteAddr(%q)=%v, want %v", tt.addr, got, tt.want)
		}
	}
}
