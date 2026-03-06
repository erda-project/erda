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
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func TestHandleAIProxyHeader(t *testing.T) {
	tests := []struct {
		name                        string
		requestID                   string
		generatedCallID             string
		backendRequestID            string
		expectedXRequestId          string
		expectedGeneratedCallId     string
		expectedLLMBackendRequestId string
	}{
		{
			name:                        "No LLM backend request ID",
			requestID:                   "client-123",
			generatedCallID:             "generated-456",
			backendRequestID:            "",
			expectedXRequestId:          "client-123",
			expectedGeneratedCallId:     "generated-456",
			expectedLLMBackendRequestId: "",
		},
		{
			name:                        "Has LLM backend request ID",
			requestID:                   "client-123",
			generatedCallID:             "generated-456",
			backendRequestID:            "backend-789",
			expectedXRequestId:          "client-123",
			expectedGeneratedCallId:     "generated-456",
			expectedLLMBackendRequestId: "backend-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest("POST", "http://example.com", nil)
			assert.NoError(t, err)

			// Create context and sync.Map, set both IDs
			ctx := ctxhelper.InitCtxMapIfNeed(req.Context())
			req = req.WithContext(ctx)
			ctxhelper.PutRequestID(ctx, tt.requestID)
			ctxhelper.PutGeneratedCallID(ctx, tt.generatedCallID)

			// Create response
			resp := &http.Response{
				Header:  make(http.Header),
				Request: req,
			}

			// If there's a backend request ID, set it in response headers
			if tt.backendRequestID != "" {
				resp.Header.Set(vars.XRequestId, tt.backendRequestID)
			}

			// Execute handleAIProxyResponseHeader
			handleAIProxyResponseHeader(resp)

			// Verify X-Request-Id header
			assert.Equal(t, tt.expectedXRequestId, resp.Header.Get(vars.XRequestId))

			// Verify X-AI-Proxy-Generated-Call-Id header
			assert.Equal(t, tt.expectedGeneratedCallId, resp.Header.Get(vars.XAIProxyGeneratedCallId))

			// Verify X-Request-Id-LLM-Backend header
			if tt.expectedLLMBackendRequestId != "" {
				assert.Equal(t, tt.expectedLLMBackendRequestId, resp.Header.Get(vars.XRequestIdLLMBackend))
			} else {
				assert.Empty(t, resp.Header.Get(vars.XRequestIdLLMBackend))
			}
		})
	}
}

func TestHandleRequestIdHeaders(t *testing.T) {
	tests := []struct {
		name                        string
		requestID                   string
		generatedCallID             string
		backendRequestID            string
		expectedXRequestId          string
		expectedGeneratedCallId     string
		expectedLLMBackendRequestId string
	}{
		{
			name:                        "Handle case without backend request ID",
			requestID:                   "client-123",
			generatedCallID:             "generated-456",
			backendRequestID:            "",
			expectedXRequestId:          "client-123",
			expectedGeneratedCallId:     "generated-456",
			expectedLLMBackendRequestId: "",
		},
		{
			name:                        "Handle case with backend request ID",
			requestID:                   "client-123",
			generatedCallID:             "generated-456",
			backendRequestID:            "llm-789",
			expectedXRequestId:          "client-123",
			expectedGeneratedCallId:     "generated-456",
			expectedLLMBackendRequestId: "llm-789",
		},
		{
			name:                        "Backend request ID same as client request ID",
			requestID:                   "same-id-999",
			generatedCallID:             "generated-456",
			backendRequestID:            "same-id-999",
			expectedXRequestId:          "same-id-999",
			expectedGeneratedCallId:     "generated-456",
			expectedLLMBackendRequestId: "same-id-999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest("POST", "http://example.com", nil)
			assert.NoError(t, err)

			// Create context and sync.Map, set both IDs
			ctx := ctxhelper.InitCtxMapIfNeed(req.Context())
			req = req.WithContext(ctx)
			ctxhelper.PutRequestID(ctx, tt.requestID)
			ctxhelper.PutGeneratedCallID(ctx, tt.generatedCallID)

			// Create response
			resp := &http.Response{
				Header:  make(http.Header),
				Request: req,
			}

			// If there's a backend request ID, set it in response headers
			if tt.backendRequestID != "" {
				resp.Header.Set(vars.XRequestId, tt.backendRequestID)
			}

			// Test the extracted method directly
			_handleRequestIdHeaders(resp)

			// Verify results
			assert.Equal(t, tt.expectedXRequestId, resp.Header.Get(vars.XRequestId))
			assert.Equal(t, tt.expectedGeneratedCallId, resp.Header.Get(vars.XAIProxyGeneratedCallId))

			if tt.expectedLLMBackendRequestId != "" {
				assert.Equal(t, tt.expectedLLMBackendRequestId, resp.Header.Get(vars.XRequestIdLLMBackend))
			} else {
				assert.Empty(t, resp.Header.Get(vars.XRequestIdLLMBackend))
			}
		})
	}
}

func TestHandleModelHealthMetaHeader(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com", nil)
	assert.NoError(t, err)

	ctx := ctxhelper.InitCtxMapIfNeed(req.Context())
	req = req.WithContext(ctx)
	ctxhelper.PutRequestID(ctx, "client-1")
	ctxhelper.PutGeneratedCallID(ctx, "call-1")
	health.AppendReleasedUnsupportedAPIType(ctx, "embeddings")
	health.AppendReleasedUnsupportedAPIType(ctx, "embeddings")

	resp := &http.Response{
		Header:  make(http.Header),
		Request: req,
	}
	handleAIProxyResponseHeader(resp)

	raw := resp.Header.Get(vars.XAIProxyModelHealthMeta)
	assert.NotEmpty(t, raw)
	assert.NotContains(t, strings.ToLower(raw), "instance_id")

	var payload map[string]any
	assert.NoError(t, json.Unmarshal([]byte(raw), &payload))
	_, hasVersion := payload["version"]
	assert.False(t, hasVersion)
	assert.EqualValues(t, 2, payload["released_unsupported_count"])
	assert.EqualValues(t, "unsupported_probe_api_type", payload["reason"])

	apiTypes, ok := payload["released_unsupported_api_types"].([]any)
	assert.True(t, ok)
	assert.Len(t, apiTypes, 1)
	assert.EqualValues(t, "embeddings", apiTypes[0])
}

func TestHandleModelMarkUnhealthyHeader(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com", nil)
	assert.NoError(t, err)

	ctx := ctxhelper.InitCtxMapIfNeed(req.Context())
	req = req.WithContext(ctx)
	ctxhelper.PutRequestID(ctx, "client-1")
	ctxhelper.PutGeneratedCallID(ctx, "call-1")
	ctxhelper.PutModelMarkUnhealthyInstanceID(ctx, "m-123")

	resp := &http.Response{
		Header:  make(http.Header),
		Request: req,
	}
	handleAIProxyResponseHeader(resp)

	assert.Equal(t, "m-123", resp.Header.Get(vars.XAIProxyModelHealthMarkUnhealthy))
}

func TestHandleModelRetryMetaHeader(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com", nil)
	assert.NoError(t, err)

	ctx := ctxhelper.InitCtxMapIfNeed(req.Context())
	req = req.WithContext(ctx)
	ctxhelper.PutRequestID(ctx, "client-1")
	ctxhelper.PutGeneratedCallID(ctx, "call-1")
	ctxhelper.PutReverseProxyRetryAttempt(ctx, 2)
	ctxhelper.PutModel(ctx, &modelpb.Model{Id: "m-2"})

	resp := &http.Response{
		Header:  make(http.Header),
		Request: req,
	}
	handleAIProxyResponseHeader(resp)

	raw := resp.Header.Get(vars.XAIProxyModelRetryMeta)
	assert.NotEmpty(t, raw)

	var payload map[string]any
	assert.NoError(t, json.Unmarshal([]byte(raw), &payload))
	assert.EqualValues(t, 2, payload["raw_llm_backend_request_count"])
	assert.EqualValues(t, 1, payload["raw_llm_backend_retry_count"])
	assert.EqualValues(t, "m-2", payload["final_model_instance_id"])
}
