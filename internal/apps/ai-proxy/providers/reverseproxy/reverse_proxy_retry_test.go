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
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	rproxy "github.com/erda-project/erda/internal/apps/ai-proxy/route/reverse_proxy"
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

	req := httptest.NewRequest(http.MethodPost, "http://ai-proxy.test/v1/chat/completions", strings.NewReader("{}"))
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
	policy := transparentRetryPolicy{
		Enabled:               true,
		MaxAttempts:           3,
		BackoffBase:           0,
		RetryableHTTPStatuses: map[int]struct{}{},
	}
	options := []OptionFunc{
		func(_ context.Context, proxy *httputil.ReverseProxy) {
			proxy.Rewrite = func(pr *httputil.ProxyRequest) {
				pr.SetURL(targetURL)
				modelID := "m-a"
				if excluded, ok := ctxhelper.GetReverseProxyRetryExcludedModelIDs(pr.In.Context()); ok {
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

	req := httptest.NewRequest(http.MethodPost, "http://ai-proxy.test/v1/chat/completions", strings.NewReader("{}"))
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	tw := rproxy.NewTrackedResponseWriter(rec)

	targetURL, err := url.Parse("http://upstream.test")
	if err != nil {
		t.Fatal(err)
	}

	attempts := 0
	ctxErrs := make([]error, 0, 2)
	policy := transparentRetryPolicy{
		Enabled:               true,
		MaxAttempts:           2,
		BackoffBase:           0,
		RetryableHTTPStatuses: map[int]struct{}{},
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
	policy := transparentRetryPolicy{
		Enabled:               true,
		MaxAttempts:           2,
		BackoffBase:           0,
		RetryableHTTPStatuses: map[int]struct{}{},
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
	policy := transparentRetryPolicy{
		Enabled:               true,
		MaxAttempts:           3,
		BackoffBase:           time.Millisecond,
		RetryableHTTPStatuses: map[int]struct{}{},
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
