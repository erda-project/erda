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

package token_usage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	usagepb "github.com/erda-project/erda-proto-go/apps/aiproxy/usage/token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func TestCalculateTokens_ExtractorPreferred(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutIsStream(ctx, false)

	respPayload := `{"usage":{"prompt_tokens":5,"completion_tokens":7,"total_tokens":12}}`
	ctxhelper.PutReverseProxyWholeHandledResponseBodyStr(ctx, respPayload)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/chat/completions", strings.NewReader(`{"input":"hello"}`))
	req = req.WithContext(ctx)
	ctxhelper.PutReverseProxyRequestInSnapshot(ctx, req)

	resp := &http.Response{Request: req}
	createReq := usagepb.TokenUsageCreateRequest{}

	calculateTokens(resp, &createReq)

	if createReq.IsEstimated {
		t.Fatalf("expected extractor path to mark IsEstimated as false")
	}
	if createReq.InputTokens != 5 || createReq.OutputTokens != 7 || createReq.TotalTokens != 12 {
		t.Fatalf("unexpected token values, got input=%d output=%d total=%d", createReq.InputTokens, createReq.OutputTokens, createReq.TotalTokens)
	}
	if createReq.UsageDetails == "" {
		t.Fatalf("expected UsageDetails to be filled by extractor")
	}
}

func TestCalculateTokens_EstimatorFallback(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutIsStream(ctx, false)
	ctxhelper.PutReverseProxyWholeHandledResponseBodyStr(ctx, `{"data":"response"}`)
	ctxhelper.PutModel(ctx, &modelpb.Model{Name: "gpt-4"})

	req := httptest.NewRequest(http.MethodPost, "http://example.com/proxy/unknown/path", strings.NewReader("prompt content"))
	req = req.WithContext(ctx)
	ctxhelper.PutReverseProxyRequestInSnapshot(ctx, req)

	resp := &http.Response{Request: req}
	createReq := usagepb.TokenUsageCreateRequest{}

	calculateTokens(resp, &createReq)

	if !createReq.IsEstimated {
		t.Fatalf("expected estimator path to mark IsEstimated as true")
	}
	if createReq.TotalTokens == 0 {
		t.Fatalf("expected estimator to populate token counts, got zero total tokens")
	}
}
