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

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	audittypes "github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func TestHandleAIProxyRequestHeaderSetsTrustedHealthProbeForLoopback(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutRequestID(ctx, "req-1")
	ctxhelper.PutGeneratedCallID(ctx, "call-1")
	ctxhelper.PutAuditSink(ctx, audittypes.New("aid-1", logrusx.New()))
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

	sink, _ := ctxhelper.GetAuditSink(ctx)
	if got := sink.Snapshot()["model_health.trusted_probe"]; got != true {
		t.Fatalf("expected trusted probe audit note, got %v", got)
	}
}

func TestHandleAIProxyRequestHeaderDoesNotSetTrustedHealthProbeForNonLoopback(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutRequestID(ctx, "req-1")
	ctxhelper.PutGeneratedCallID(ctx, "call-1")
	ctxhelper.PutAuditSink(ctx, audittypes.New("aid-1", logrusx.New()))
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8081/v1/chat/completions", nil).WithContext(ctx)
	req.RemoteAddr = "10.20.30.40:43123"
	req.Header.Set(vars.XAIProxyModelHealthProbe, "true")
	out := req.Clone(ctx)
	pr := &httputil.ProxyRequest{In: req, Out: out}

	handleAIProxyRequestHeader(pr)

	if trusted, ok := ctxhelper.GetTrustedHealthProbe(ctx); ok && trusted {
		t.Fatal("did not expect trusted health probe for non-loopback request")
	}

	sink, _ := ctxhelper.GetAuditSink(ctx)
	if _, ok := sink.Snapshot()["model_health.trusted_probe"]; ok {
		t.Fatal("did not expect trusted probe audit note for non-loopback request")
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

func TestNoteRetryAuditMetadata(t *testing.T) {
	cases := []struct {
		name              string
		attempt           int
		wantRequestCount  any
		wantRetryCount    any
		expectRetryFields bool
	}{
		{name: "first attempt", attempt: 1, expectRetryFields: false},
		{name: "second attempt", attempt: 2, wantRequestCount: 2, wantRetryCount: 1, expectRetryFields: true},
		{name: "third attempt", attempt: 3, wantRequestCount: 3, wantRetryCount: 2, expectRetryFields: true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
			ctxhelper.PutAuditSink(ctx, audittypes.New("aid-1", logrusx.New()))
			ctxhelper.PutModelRetryRawLLMBackendRequestCount(ctx, tt.attempt)

			noteRetryAuditMetadata(ctx)

			sink, _ := ctxhelper.GetAuditSink(ctx)
			got := sink.Snapshot()
			if !tt.expectRetryFields {
				if _, ok := got["model_retry.raw_llm_backend_request_count"]; ok {
					t.Fatalf("did not expect request_count audit field on attempt=%d", tt.attempt)
				}
				if _, ok := got["model_retry.raw_llm_backend_retry_count"]; ok {
					t.Fatalf("did not expect retry_count audit field on attempt=%d", tt.attempt)
				}
			} else {
				if got["model_retry.raw_llm_backend_request_count"] != tt.wantRequestCount {
					t.Fatalf("unexpected request_count, got=%v want=%v", got["model_retry.raw_llm_backend_request_count"], tt.wantRequestCount)
				}
				if got["model_retry.raw_llm_backend_retry_count"] != tt.wantRetryCount {
					t.Fatalf("unexpected retry_count, got=%v want=%v", got["model_retry.raw_llm_backend_retry_count"], tt.wantRetryCount)
				}
			}
			if _, ok := got["reverse_proxy.retry.final_llm_backend_request_count"]; ok {
				t.Fatal("did not expect legacy final_llm_backend_request_count audit field")
			}
			if _, ok := got["reverse_proxy.retry.attempts"]; ok {
				t.Fatal("did not expect legacy retry.attempts audit field")
			}
			if _, ok := got["reverse_proxy.retry.final_instance_id"]; ok {
				t.Fatal("did not expect legacy final_instance_id audit field")
			}
		})
	}
}
