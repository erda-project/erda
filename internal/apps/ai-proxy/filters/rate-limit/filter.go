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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/erda-project/erda-infra/base/logs"
	auditpb "github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "rate-limit"
)

var (
	_ reverseproxy.RequestFilter  = (*RateLimiter)(nil)
	_ reverseproxy.ResponseFilter = (*RateLimiter)(nil)
)

type TokenLimiter struct {
	mu      sync.Mutex
	limiter map[string]*rate.Limiter
}

var tokenLimiter *TokenLimiter

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)

	// init pkg level vars
	tokenLimiter = &TokenLimiter{
		mu:      sync.Mutex{},
		limiter: make(map[string]*rate.Limiter),
	}
}

type RateLimiter struct {
	*reverseproxy.DefaultResponseFilter
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	f := &RateLimiter{
		DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter(),
	}
	return f, nil
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

func (f *RateLimiter) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)
	l.Sub(Name)

	// only limit token rate now
	token, isTokenInvoke := isClientTokenInvoke(infor)
	if !isTokenInvoke {
		return reverseproxy.Continue, nil
	}

	limiter := tokenLimiter.GetLimiter(token)
	if !limiter.Allow() {
		err := fmt.Errorf("too many requests for token rate limit, token: %s", token)
		l.Warn(err)
		http.Error(w, "too many requests for token rate limit", http.StatusTooManyRequests)
		// record to audit
		auditID, ok := ctxhelper.GetAuditID(ctx)
		if ok {
			_, err := ctxhelper.MustGetDBClient(ctx).AuditClient().SetFilterErrorAudit(ctx, &auditpb.AuditUpdateRequestWhenFilterError{
				AuditId:     auditID,
				FilterName:  Name,
				FilterError: err.Error(),
			})
			if err != nil {
				l := ctxhelper.GetLogger(ctx)
				l.Errorf("failed to update audit when rate limited, audit id: %s, err: %v", auditID, err)
			}
		}
		return reverseproxy.Intercept, nil
	}
	l.Debugf("pass token rate limit, token: %s", token)

	return reverseproxy.Continue, nil
}

func isClientTokenInvoke(infor reverseproxy.HttpInfor) (string, bool) {
	authHeader := vars.TrimBearer(infor.Header().Get(httputil.HeaderKeyAuthorization))
	if strings.HasPrefix(authHeader, client_token.TokenPrefix) {
		return authHeader, true
	}
	return "", false
}
