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

package model_retry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func TestWithRequestOverrides(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLoggerBase(ctx, logrusx.New())

	req := httptest.NewRequest(http.MethodPost, "http://ai-proxy.test/v1/chat/completions", nil).WithContext(ctx)
	req.Header.Set(vars.XAIProxyModelRetry, "false")
	req.Header.Set(vars.XAIProxyModelRetryMax, "5")

	policy := (Config{
		Enabled: true,
		Conditions: Conditions{
			MaxLLMBackendRequestCount: 3,
			Backoff: Backoff{
				Base: time.Second,
				Max:  10 * time.Second,
			},
			RetryableHTTPStatuses: []int{429, 503},
		},
		Actions: Actions{
			ExcludeFailedInstance: true,
		},
		Observability: Observability{
			ResponseHeaderMeta: true,
		},
	}).WithRequestOverrides(req)

	if policy.Enabled {
		t.Fatal("expected retry disabled by request header")
	}
	if policy.Conditions.MaxLLMBackendRequestCount != 5 {
		t.Fatalf("expected max request count=5, got %d", policy.Conditions.MaxLLMBackendRequestCount)
	}
	if !policy.IsRetryableHTTPStatus(429) {
		t.Fatal("expected 429 to be retryable")
	}
}

func TestWithRequestOverridesDisablesRetryForHealthProbe(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLoggerBase(ctx, logrusx.New())

	req := httptest.NewRequest(http.MethodPost, "http://ai-proxy.test/v1/chat/completions", nil).WithContext(ctx)
	req.Header.Set(vars.XAIProxyModelHealthProbe, "true")

	policy := (Config{Enabled: true, Conditions: Conditions{MaxLLMBackendRequestCount: 3}}).WithRequestOverrides(req)
	if policy.Enabled {
		t.Fatal("expected health probe request to disable retry")
	}
}

func TestExcludedModelIDs(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())

	AddExcludedModelID(ctx, "m-1")
	AddExcludedModelID(ctx, "m-2")

	got, ok := GetExcludedModelIDs(ctx)
	if !ok {
		t.Fatal("expected excluded model ids to exist")
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 excluded model ids, got %d", len(got))
	}
	if _, ok := got["m-1"]; !ok {
		t.Fatal("expected m-1 to be excluded")
	}
	if _, ok := got["m-2"]; !ok {
		t.Fatal("expected m-2 to be excluded")
	}
}
