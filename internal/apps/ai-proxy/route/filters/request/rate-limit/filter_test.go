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

package rate_limit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"golang.org/x/time/rate"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func TestRateLimiter_BypassHealthProbe(t *testing.T) {
	resetTokenLimiterForTest()

	limiter := &RateLimiter{}
	proxyReq := newProxyRequestForTest()
	ctxhelper.PutTrustedHealthProbe(proxyReq.In.Context(), true)

	for i := 0; i < 10; i++ {
		if err := limiter.OnProxyRequest(proxyReq); err != nil {
			t.Fatalf("probe request should bypass rate limit, got err at %d: %v", i, err)
		}
	}
}

func TestRateLimiter_ProbeHeaderWithoutTrustedContextShouldNotBypass(t *testing.T) {
	resetTokenLimiterForTest()

	limiter := &RateLimiter{}
	proxyReq := newProxyRequestForTest()
	proxyReq.In.Header.Set(vars.XAIProxyHealthProbe, "true")

	if err := limiter.OnProxyRequest(proxyReq); err != nil {
		t.Fatalf("first request should pass, got %v", err)
	}
	if err := limiter.OnProxyRequest(proxyReq); err != nil {
		t.Fatalf("second request should pass, got %v", err)
	}
	if err := limiter.OnProxyRequest(proxyReq); err == nil {
		t.Fatal("third request should be rate limited without trusted probe context")
	}
}

func TestRateLimiter_StillLimitNormalRequest(t *testing.T) {
	resetTokenLimiterForTest()

	limiter := &RateLimiter{}
	proxyReq := newProxyRequestForTest()

	if err := limiter.OnProxyRequest(proxyReq); err != nil {
		t.Fatalf("first request should pass, got %v", err)
	}
	if err := limiter.OnProxyRequest(proxyReq); err != nil {
		t.Fatalf("second request should pass, got %v", err)
	}
	if err := limiter.OnProxyRequest(proxyReq); err == nil {
		t.Fatal("third request should be rate limited")
	}
}

func newProxyRequestForTest() *httputil.ProxyRequest {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLogger(ctx, logrusx.New())

	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/chat/completions", nil).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+client_token.TokenPrefix+"token")
	outReq := req.Clone(ctx)

	return &httputil.ProxyRequest{
		In:  req,
		Out: outReq,
	}
}

func resetTokenLimiterForTest() {
	tokenLimiter = &TokenLimiter{
		limiter: make(map[string]*rate.Limiter),
	}
}
