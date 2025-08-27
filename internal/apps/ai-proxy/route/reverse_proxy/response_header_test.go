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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
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

			// Execute handleAIProxyHeader
			handleAIProxyHeader(resp)

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
