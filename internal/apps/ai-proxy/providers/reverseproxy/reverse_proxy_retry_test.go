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

package reverseproxy

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	audittypes "github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	modelretry "github.com/erda-project/erda/internal/apps/ai-proxy/providers/reverseproxy/retry/model_retry"
	httperror "github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	rproxy "github.com/erda-project/erda/internal/apps/ai-proxy/route/reverse_proxy"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type flakyStreamBody struct {
	emitted bool
}

func (b *flakyStreamBody) Read(p []byte) (int, error) {
	if !b.emitted {
		b.emitted = true
		copy(p, "chunk")
		return len("chunk"), nil
	}
	return 0, errors.New("read tcp 127.0.0.1:1->127.0.0.1:2: read: connection reset by peer")
}

func (b *flakyStreamBody) Close() error { return nil }

func TestServeWithTransparentRetry_FirstNetworkFailureThenSuccess(t *testing.T) {
	p := &provider{}

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLogger(ctx, logrusx.New())
	ctxhelper.PutLoggerBase(ctx, logrusx.New())
	ctxhelper.PutReverseProxyRequestBodyBytes(ctx, []byte("{}"))
	ctxhelper.PutRequestID(ctx, "req-429")
	ctxhelper.PutGeneratedCallID(ctx, "call-429")

	req := httptest.NewRequest(http.MethodPost, "http://ai-proxy.test/v1/chat/completions", strings.NewReader("{}"))
	req.Header.Set(vars.XRequestId, "req-429")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	tw := rproxy.NewTrackedResponseWriter(rec)

	targetURL, err := url.Parse("http://upstream.test")
	if err != nil {
		t.Fatal(err)
	}

	attempts := 0
	var selectedInstances []string
	var requestBodies []string
	policy := modelretry.Config{
		Enabled: true,
		Conditions: modelretry.Conditions{
			MaxLLMBackendRequestCount: 3,
			Backoff:                   modelretry.Backoff{Base: 0},
			RetryableHTTPStatuses:     nil,
		},
		Actions: modelretry.Actions{ExcludeFailedInstance: true},
	}
	options := []OptionFunc{
		func(_ context.Context, proxy *httputil.ReverseProxy) {
			proxy.Rewrite = func(pr *httputil.ProxyRequest) {
				pr.SetURL(targetURL)
				modelID := "m-a"
				if excluded, ok := modelretry.GetExcludedModelIDs(pr.In.Context()); ok {
					if _, hit := excluded[modelID]; hit {
						modelID = "m-b"
					}
				}
				selectedInstances = append(selectedInstances, modelID)
				ctxhelper.PutModel(pr.In.Context(), &modelpb.Model{Id: modelID})
			}
			proxy.ModifyResponse = nil
			proxy.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				bodyBytes, _ := io.ReadAll(req.Body)
				requestBodies = append(requestBodies, string(bodyBytes))
				if attempts == 1 {
					return nil, errors.New("dial tcp 127.0.0.1:80: connect: connection refused")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
					Request:    req,
				}, nil
			})
		},
	}

	p.serveWithTransparentRetry(ctx, tw, req, nil, nil, options, policy)

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Fatalf("expected single successful body, got %q", got)
	}
	if len(requestBodies) != 2 || requestBodies[0] != "{}" || requestBodies[1] != "{}" {
		t.Fatalf("expected request body to be reset for each attempt, got: %v", requestBodies)
	}
	if len(selectedInstances) != 2 || selectedInstances[0] != "m-a" || selectedInstances[1] != "m-b" {
		t.Fatalf("expected retry to avoid failed instance, got: %v", selectedInstances)
	}
}

func TestTransparentRetry_ContextNotCanceledAcrossAttempts(t *testing.T) {
	p := &provider{}

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLogger(ctx, logrusx.New())
	ctxhelper.PutLoggerBase(ctx, logrusx.New())
	ctxhelper.PutReverseProxyRequestBodyBytes(ctx, []byte("{}"))
	ctxhelper.PutRequestID(ctx, "req-400")
	ctxhelper.PutGeneratedCallID(ctx, "call-400")

	req := httptest.NewRequest(http.MethodPost, "http://ai-proxy.test/v1/chat/completions", strings.NewReader("{}"))
	req.Header.Set(vars.XRequestId, "req-400")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	tw := rproxy.NewTrackedResponseWriter(rec)

	targetURL, err := url.Parse("http://upstream.test")
	if err != nil {
		t.Fatal(err)
	}

	attempts := 0
	ctxErrs := make([]error, 0, 2)
	policy := modelretry.Config{
		Enabled: true,
		Conditions: modelretry.Conditions{
			MaxLLMBackendRequestCount: 2,
			Backoff:                   modelretry.Backoff{Base: 0},
			RetryableHTTPStatuses:     nil,
		},
		Actions: modelretry.Actions{ExcludeFailedInstance: true},
	}
	options := []OptionFunc{
		func(_ context.Context, proxy *httputil.ReverseProxy) {
			proxy.Rewrite = func(pr *httputil.ProxyRequest) {
				pr.SetURL(targetURL)
			}
			proxy.ModifyResponse = nil
			proxy.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				ctxErrs = append(ctxErrs, req.Context().Err())
				if attempts == 1 {
					return nil, errors.New("dial tcp 127.0.0.1:80: connect: connection refused")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
					Request:    req,
				}, nil
			})
		},
	}

	p.serveWithTransparentRetry(ctx, tw, req, nil, nil, options, policy)

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if len(ctxErrs) != 2 {
		t.Fatalf("expected context errs captured for both attempts, got %d", len(ctxErrs))
	}
	if ctxErrs[1] != nil {
		t.Fatalf("expected second attempt context err to be nil, got %v", ctxErrs[1])
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Fatalf("expected retry path to continue and return success body, got %q", got)
	}
}

func TestTransparentRetry_ReusesRequestIDButRegeneratesCallID(t *testing.T) {
	p := &provider{}

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLogger(ctx, logrusx.New())
	ctxhelper.PutLoggerBase(ctx, logrusx.New())
	ctxhelper.PutReverseProxyRequestBodyBytes(ctx, []byte("{}"))
	ctxhelper.PutRequestID(ctx, "req-1")
	ctxhelper.PutGeneratedCallID(ctx, "call-1")

	req := httptest.NewRequest(http.MethodPost, "http://ai-proxy.test/v1/chat/completions", strings.NewReader("{}"))
	req.Header.Set(vars.XRequestId, "req-1")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	tw := rproxy.NewTrackedResponseWriter(rec)

	targetURL, err := url.Parse("http://upstream.test")
	if err != nil {
		t.Fatal(err)
	}

	attempts := 0
	var requestIDs []string
	var callIDs []string
	policy := modelretry.Config{
		Enabled: true,
		Conditions: modelretry.Conditions{
			MaxLLMBackendRequestCount: 2,
			Backoff:                   modelretry.Backoff{Base: 0},
			RetryableHTTPStatuses:     nil,
		},
		Actions: modelretry.Actions{ExcludeFailedInstance: true},
	}
	options := []OptionFunc{
		func(_ context.Context, proxy *httputil.ReverseProxy) {
			proxy.Rewrite = func(pr *httputil.ProxyRequest) {
				pr.SetURL(targetURL)
				requestIDs = append(requestIDs, ctxhelper.MustGetRequestID(pr.In.Context()))
				callIDs = append(callIDs, ctxhelper.MustGetGeneratedCallID(pr.In.Context()))
			}
			proxy.ModifyResponse = nil
			proxy.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				if attempts == 1 {
					return nil, errors.New("dial tcp 127.0.0.1:80: connect: connection refused")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
					Request:    req,
				}, nil
			})
		},
	}

	p.serveWithTransparentRetry(ctx, tw, req, nil, nil, options, policy)

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if len(requestIDs) != 2 || requestIDs[0] != "req-1" || requestIDs[1] != "req-1" {
		t.Fatalf("expected same request id for all attempts, got %v", requestIDs)
	}
	if len(callIDs) != 2 {
		t.Fatalf("expected call ids for both attempts, got %v", callIDs)
	}
	if callIDs[0] != "call-1" {
		t.Fatalf("expected first attempt keep initial call id, got %q", callIDs[0])
	}
	if callIDs[1] == "call-1" {
		t.Fatalf("expected retry attempt to regenerate call id, got %v", callIDs)
	}
	if callIDs[0] == callIDs[1] {
		t.Fatalf("expected different call ids across attempts, got %v", callIDs)
	}
}

func TestServeWithTransparentRetry_NoRetryAfterHeadersWritten(t *testing.T) {
	p := &provider{}

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLogger(ctx, logrusx.New())
	ctxhelper.PutLoggerBase(ctx, logrusx.New())

	req := httptest.NewRequest(http.MethodGet, "http://ai-proxy.test/v1/chat/completions", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	tw := rproxy.NewTrackedResponseWriter(rec)

	targetURL, err := url.Parse("http://upstream.test")
	if err != nil {
		t.Fatal(err)
	}

	attempts := 0
	policy := modelretry.Config{
		Enabled: true,
		Conditions: modelretry.Conditions{
			MaxLLMBackendRequestCount: 3,
			Backoff:                   modelretry.Backoff{Base: time.Millisecond},
			RetryableHTTPStatuses:     nil,
		},
		Actions: modelretry.Actions{ExcludeFailedInstance: true},
	}
	options := []OptionFunc{
		func(_ context.Context, proxy *httputil.ReverseProxy) {
			proxy.Rewrite = func(pr *httputil.ProxyRequest) {
				pr.SetURL(targetURL)
			}
			proxy.ModifyResponse = nil
			proxy.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       &flakyStreamBody{},
					Request:    req,
				}, nil
			})
		},
	}

	p.serveWithTransparentRetry(ctx, tw, req, nil, nil, options, policy)

	if attempts != 1 {
		t.Fatalf("expected no retry once headers/body have started, got attempts=%d", attempts)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got == "" {
		t.Fatalf("expected streamed body bytes before failure")
	}
}

func TestServeWithTransparentRetry_ExcludeFailedInstanceDisabled(t *testing.T) {
	p := &provider{}

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLogger(ctx, logrusx.New())
	ctxhelper.PutLoggerBase(ctx, logrusx.New())
	ctxhelper.PutReverseProxyRequestBodyBytes(ctx, []byte("{}"))
	ctxhelper.PutRequestID(ctx, "req-400")
	ctxhelper.PutGeneratedCallID(ctx, "call-400")

	req := httptest.NewRequest(http.MethodPost, "http://ai-proxy.test/v1/chat/completions", strings.NewReader("{}"))
	req.Header.Set(vars.XRequestId, "req-400")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	tw := rproxy.NewTrackedResponseWriter(rec)

	targetURL, err := url.Parse("http://upstream.test")
	if err != nil {
		t.Fatal(err)
	}

	attempts := 0
	var selectedInstances []string
	policy := modelretry.Config{
		Enabled: true,
		Conditions: modelretry.Conditions{
			MaxLLMBackendRequestCount: 2,
			Backoff:                   modelretry.Backoff{Base: 0},
			RetryableHTTPStatuses:     nil,
		},
		Actions: modelretry.Actions{ExcludeFailedInstance: false},
	}
	options := []OptionFunc{
		func(_ context.Context, proxy *httputil.ReverseProxy) {
			proxy.Rewrite = func(pr *httputil.ProxyRequest) {
				pr.SetURL(targetURL)
				modelID := "m-a"
				if excluded, ok := modelretry.GetExcludedModelIDs(pr.In.Context()); ok {
					if _, hit := excluded[modelID]; hit {
						modelID = "m-b"
					}
				}
				selectedInstances = append(selectedInstances, modelID)
				ctxhelper.PutModel(pr.In.Context(), &modelpb.Model{Id: modelID})
			}
			proxy.ModifyResponse = nil
			proxy.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				if attempts == 1 {
					return nil, errors.New("dial tcp 127.0.0.1:80: connect: connection refused")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
					Request:    req,
				}, nil
			})
		},
	}

	p.serveWithTransparentRetry(ctx, tw, req, nil, nil, options, policy)

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if len(selectedInstances) != 2 || selectedInstances[0] != "m-a" || selectedInstances[1] != "m-a" {
		t.Fatalf("expected retry to allow the same instance when exclusion is disabled, got: %v", selectedInstances)
	}
}

func TestNextRetryBackoff_RespectsMax(t *testing.T) {
	policy := modelretry.Config{
		Conditions: modelretry.Conditions{
			Backoff: modelretry.Backoff{
				Base: time.Second,
				Max:  4 * time.Second,
			},
		},
	}

	if got := policy.NextBackoff(3); got != 4*time.Second {
		t.Fatalf("expected backoff capped at 4s, got %s", got)
	}
}

func TestServeWithTransparentRetry_RetryOnHTTP429(t *testing.T) {
	p := &provider{}

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLogger(ctx, logrusx.New())
	ctxhelper.PutLoggerBase(ctx, logrusx.New())
	ctxhelper.PutReverseProxyRequestBodyBytes(ctx, []byte("{}"))
	ctxhelper.PutRequestID(ctx, "req-400")
	ctxhelper.PutGeneratedCallID(ctx, "call-400")

	req := httptest.NewRequest(http.MethodPost, "http://ai-proxy.test/v1/chat/completions", strings.NewReader("{}"))
	req.Header.Set(vars.XRequestId, "req-400")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	tw := rproxy.NewTrackedResponseWriter(rec)

	targetURL, err := url.Parse("http://upstream.test")
	if err != nil {
		t.Fatal(err)
	}

	attempts := 0
	var selectedInstances []string
	policy := modelretry.Config{
		Enabled: true,
		Conditions: modelretry.Conditions{
			MaxLLMBackendRequestCount: 2,
			Backoff:                   modelretry.Backoff{Base: 0},
			RetryableHTTPStatuses:     []int{http.StatusTooManyRequests},
		},
		Actions: modelretry.Actions{ExcludeFailedInstance: true},
	}
	options := []OptionFunc{
		func(_ context.Context, proxy *httputil.ReverseProxy) {
			proxy.Rewrite = func(pr *httputil.ProxyRequest) {
				pr.SetURL(targetURL)
				modelID := "m-a"
				if excluded, ok := modelretry.GetExcludedModelIDs(pr.In.Context()); ok {
					if _, hit := excluded[modelID]; hit {
						modelID = "m-b"
					}
				}
				selectedInstances = append(selectedInstances, modelID)
				ctxhelper.PutModel(pr.In.Context(), &modelpb.Model{Id: modelID})
			}
			proxy.ModifyResponse = nil
			proxy.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				if attempts == 1 {
					return nil, httperror.NewHTTPError(req.Context(), http.StatusTooManyRequests, "rate limited")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("ok")),
					Request:    req,
				}, nil
			})
		},
	}

	p.serveWithTransparentRetry(ctx, tw, req, nil, nil, options, policy)

	if attempts != 2 {
		t.Fatalf("expected 2 attempts for retryable 429, got %d", attempts)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected final status 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Fatalf("expected successful retry body, got %q", got)
	}
	if len(selectedInstances) != 2 || selectedInstances[0] != "m-a" || selectedInstances[1] != "m-b" {
		t.Fatalf("expected retry after 429 to avoid failed instance, got: %v", selectedInstances)
	}
}

func TestServeWithTransparentRetry_DoNotRetryOnHTTP400(t *testing.T) {
	p := &provider{}

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLogger(ctx, logrusx.New())
	ctxhelper.PutLoggerBase(ctx, logrusx.New())
	ctxhelper.PutReverseProxyRequestBodyBytes(ctx, []byte("{}"))
	ctxhelper.PutRequestID(ctx, "req-400")
	ctxhelper.PutGeneratedCallID(ctx, "call-400")

	req := httptest.NewRequest(http.MethodPost, "http://ai-proxy.test/v1/chat/completions", strings.NewReader("{}"))
	req.Header.Set(vars.XRequestId, "req-400")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	tw := rproxy.NewTrackedResponseWriter(rec)

	targetURL, err := url.Parse("http://upstream.test")
	if err != nil {
		t.Fatal(err)
	}

	attempts := 0
	policy := modelretry.Config{
		Enabled: true,
		Conditions: modelretry.Conditions{
			MaxLLMBackendRequestCount: 2,
			Backoff:                   modelretry.Backoff{Base: 0},
			RetryableHTTPStatuses:     []int{http.StatusTooManyRequests},
		},
		Actions: modelretry.Actions{ExcludeFailedInstance: true},
	}
	options := []OptionFunc{
		func(_ context.Context, proxy *httputil.ReverseProxy) {
			proxy.Rewrite = func(pr *httputil.ProxyRequest) {
				pr.SetURL(targetURL)
				ctxhelper.PutModel(pr.In.Context(), &modelpb.Model{Id: "m-a"})
			}
			proxy.ModifyResponse = nil
			proxy.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				return nil, httperror.NewHTTPError(req.Context(), http.StatusBadRequest, "bad request")
			})
		},
	}

	p.serveWithTransparentRetry(ctx, tw, req, nil, nil, options, policy)

	if attempts != 1 {
		t.Fatalf("expected no retry for non-retryable 400, got %d attempts", attempts)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected final status 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "bad request") {
		t.Fatalf("expected error body to contain original message, got %q", rec.Body.String())
	}
}

func TestDisabledModelRetryDoesNotWriteRetryMetadata(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLogger(ctx, logrusx.New())
	ctxhelper.PutLoggerBase(ctx, logrusx.New())
	ctxhelper.PutAuditSink(ctx, audittypes.New("aid-1", logrusx.New()))

	policy := modelretry.Config{
		Enabled: false,
		Conditions: modelretry.Conditions{
			MaxLLMBackendRequestCount: 3,
			Backoff: modelretry.Backoff{
				Base: time.Second,
				Max:  10 * time.Second,
			},
			RetryableHTTPStatuses: []int{http.StatusTooManyRequests},
		},
		Actions: modelretry.Actions{
			ExcludeFailedInstance: true,
		},
		Observability: modelretry.Observability{
			ResponseHeaderMeta: true,
		},
	}

	if policy.Enabled {
		noteModelRetryPolicy(ctx, policy)
	}

	if _, ok := ctxhelper.GetModelRetryResponseHeaderMetaEnabled(ctx); ok {
		t.Fatal("did not expect retry response header meta flag when retry is disabled")
	}

	sink, ok := ctxhelper.GetAuditSink(ctx)
	if !ok || sink == nil {
		t.Fatal("expected audit sink")
	}
	got := sink.Snapshot()
	for _, key := range []string{
		"reverse_proxy.retry.enabled",
		"reverse_proxy.retry.max_llm_backend_request_count",
		"reverse_proxy.retry.backoff_base",
		"reverse_proxy.retry.backoff_max",
		"reverse_proxy.retry.retryable_http_statuses",
		"reverse_proxy.retry.match_network_issue_from_response_body",
		"reverse_proxy.retry.exclude_failed_instance",
		"reverse_proxy.retry.response_header_meta",
	} {
		if _, exists := got[key]; exists {
			t.Fatalf("did not expect retry audit field %s when retry is disabled", key)
		}
	}
}
