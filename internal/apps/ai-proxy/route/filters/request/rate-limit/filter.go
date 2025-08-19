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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

const name = "rate-limit"

var (
	_ filter_define.ProxyRequestRewriter = (*RateLimiter)(nil)
)

type TokenLimiter struct {
	mu      sync.Mutex
	limiter map[string]*rate.Limiter
}

var tokenLimiter *TokenLimiter

func init() {
	filter_define.RegisterFilterCreator(name, Creator)

	// init pkg level vars
	tokenLimiter = &TokenLimiter{
		mu:      sync.Mutex{},
		limiter: make(map[string]*rate.Limiter),
	}
}

type RateLimiter struct{}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &RateLimiter{}
}

func (ul *TokenLimiter) GetLimiter(token string) *rate.Limiter {
	ul.mu.Lock()
	defer ul.mu.Unlock()

	limiter, ok := ul.limiter[token]
	if !ok {
		ul.limiter[token] = rate.NewLimiter(rate.Every(time.Second), 2) // burst = 2
		return ul.limiter[token]
	}

	return limiter
}

func (f *RateLimiter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	ctx := pr.In.Context()
	l := ctxhelper.MustGetLogger(ctx)

	// only limit token rate now
	token, isTokenInvoke := isClientTokenInvoke(pr)
	if !isTokenInvoke {
		return nil
	}

	limiter := tokenLimiter.GetLimiter(token)
	if !limiter.Allow() {
		err := fmt.Errorf("too many requests for token rate limit, token: %s", token)
		l.Warn(err)
		return http_error.NewHTTPError(http.StatusTooManyRequests, "too many requests for token rate limit")
	}
	l.Debugf("pass token rate limit, token: %s", token)

	return nil
}

func isClientTokenInvoke(pr *httputil.ProxyRequest) (string, bool) {
	authHeader := vars.TrimBearer(pr.In.Header.Get(httperrorutil.HeaderKeyAuthorization))
	if strings.HasPrefix(authHeader, client_token.TokenPrefix) {
		return authHeader, true
	}
	return "", false
}
