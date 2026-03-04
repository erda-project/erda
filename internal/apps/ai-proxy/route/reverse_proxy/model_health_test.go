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
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func TestIsNetworkFailureError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "connection reset",
			err:  errors.New("read tcp 127.0.0.1:12345->127.0.0.1:443: read: connection reset by peer"),
			want: true,
		},
		{
			name: "broken pipe",
			err:  errors.New("write tcp 127.0.0.1:12345->127.0.0.1:443: write: broken pipe"),
			want: true,
		},
		{
			name: "url timeout",
			err: &url.Error{
				Op:  "Get",
				URL: "https://example.com",
				Err: errors.New("i/o timeout"),
			},
			want: true,
		},
		{
			name: "context canceled",
			err:  context.Canceled,
			want: false,
		},
		{
			name: "non network",
			err:  errors.New("invalid request body"),
			want: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := health.IsNetworkFailureError(tt.err); got != tt.want {
				t.Fatalf("IsNetworkFailureError(%v)=%v, want=%v", tt.err, got, tt.want)
			}
		})
	}
}

func TestReportModelNetworkFailure(t *testing.T) {
	manager := health.NewManager(state_store.NewMemoryStateStore(), health.Config{
		Probe: health.ProbeConfig{
			BaseURL:      "http://127.0.0.1:65530",
			UnhealthyTTL: time.Hour,
			Timeout:      2 * time.Second,
		},
		Rescue: health.RescueConfig{
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     50 * time.Millisecond,
		},
	})
	health.SetManager(manager)
	defer health.SetManager(nil)

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutModel(ctx, &modelpb.Model{Id: "m-chat"})
	req := httptest.NewRequest(http.MethodPost, vars.RequestPathPrefixV1ChatCompletions, nil).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer t_chat")
	ctxhelper.PutReverseProxyRequestInSnapshot(ctx, req)

	health.ReportModelNetworkFailure(ctx, req, errors.New("read tcp 127.0.0.1:1->127.0.0.1:2: read: connection reset by peer"))

	waitForCondition(t, 2*time.Second, func() bool {
		state, ok, _ := manager.GetState(context.Background(), "m-chat")
		return ok && state != nil
	})

	state, ok, err := manager.GetState(context.Background(), "m-chat")
	if err != nil {
		t.Fatalf("get health state failed: %v", err)
	}
	if !ok || state == nil {
		t.Fatal("expected health state for m-chat")
	}
	if state.APIType != health.APITypeChatCompletions {
		t.Fatalf("expected api_type chat_completions, got %s", state.APIType)
	}
}

func TestReportModelNetworkFailure_ProbeAndPathGuard(t *testing.T) {
	manager := health.NewManager(state_store.NewMemoryStateStore(), health.Config{
		Probe: health.ProbeConfig{
			BaseURL:      "http://127.0.0.1:65530",
			UnhealthyTTL: time.Hour,
			Timeout:      100 * time.Millisecond,
		},
		Rescue: health.RescueConfig{
			InitialBackoff: 50 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
		},
	})
	health.SetManager(manager)
	defer health.SetManager(nil)

	ctxProbe := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutModel(ctxProbe, &modelpb.Model{Id: "m-probe"})
	ctxhelper.PutTrustedHealthProbe(ctxProbe, true)
	probeReq := httptest.NewRequest(http.MethodPost, vars.RequestPathPrefixV1ChatCompletions, nil).WithContext(ctxProbe)
	probeReq.Header.Set(vars.XAIProxyHealthProbe, "true")
	ctxhelper.PutReverseProxyRequestInSnapshot(ctxProbe, probeReq)
	health.ReportModelNetworkFailure(ctxProbe, probeReq, errors.New("read tcp x->y: read: connection reset by peer"))

	if _, ok, _ := manager.GetState(context.Background(), "m-probe"); ok {
		t.Fatal("probe request should not report unhealthy")
	}

	ctxOtherPath := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutModel(ctxOtherPath, &modelpb.Model{Id: "m-other"})
	otherPathReq := httptest.NewRequest(http.MethodPost, vars.RequestPathPrefixV1Embeddings, nil).WithContext(ctxOtherPath)
	ctxhelper.PutReverseProxyRequestInSnapshot(ctxOtherPath, otherPathReq)
	health.ReportModelNetworkFailure(ctxOtherPath, otherPathReq, errors.New("read tcp x->y: read: connection reset by peer"))

	if _, ok, _ := manager.GetState(context.Background(), "m-other"); ok {
		t.Fatal("non chat/responses request should not report unhealthy")
	}
}

func waitForCondition(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("condition not met in %s", timeout)
}
