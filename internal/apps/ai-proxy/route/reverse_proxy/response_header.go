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

	"github.com/labstack/echo"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	custom_http_director "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/custom-http-director"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func handleAIProxyResponseHeader(resp *http.Response) {
	// call all header handling functions
	_handleModelHeaders(resp)
	_handleRequestIdHeaders(resp)
	_handleRequestBodyTransformHeaders(resp)
	_handleRequestThinkingTransformHeaders(resp)
	_handleEnsureContentType(resp)
	_handleContentLength(resp)
	_handleCORS(resp)
}

// _handleModelHeaders handles model related header settings
func _handleModelHeaders(resp *http.Response) {
	if model, ok := ctxhelper.GetModel(resp.Request.Context()); ok && model != nil {
		resp.Header.Set(vars.XAIProxyModelId, model.Id)
		resp.Header.Set(vars.XAIProxyModelName, model.Name)
		if provider, ok := ctxhelper.GetServiceProvider(resp.Request.Context()); ok && provider != nil {
			resp.Header.Set(vars.XAIProxyProviderName, provider.Name)
		}
	}
}

// _handleRequestIdHeaders handles request ID related header settings
func _handleRequestIdHeaders(resp *http.Response) {
	// handle X-Request-Id returned by LLM backend, rename to X-Request-Id-LLM-Backend
	if backendRequestID := resp.Header.Get(vars.XRequestId); backendRequestID != "" {
		resp.Header.Set(vars.XRequestIdLLMBackend, backendRequestID)
		// delete original X-Request-Id to avoid conflicts
		resp.Header.Del(vars.XRequestId)
	}

	// set two independent IDs to response headers
	resp.Header.Set(vars.XRequestId, ctxhelper.MustGetRequestID(resp.Request.Context()))                    // client-controllable Request ID
	resp.Header.Set(vars.XAIProxyGeneratedCallId, ctxhelper.MustGetGeneratedCallID(resp.Request.Context())) // system-generated Call ID
}

// _handleRequestBodyTransformHeaders handles body transformation related headers
func _handleRequestBodyTransformHeaders(resp *http.Response) {
	bodyTransformChanges, ok := ctxhelper.GetRequestBodyTransformChanges(resp.Request.Context())
	if !ok || bodyTransformChanges == nil {
		return
	}
	v, ok := bodyTransformChanges.(*[]custom_http_director.BodyTransformChange)
	if !ok {
		return
	}
	if changesJSON, err := json.Marshal(v); err == nil {
		resp.Header.Set(vars.XAIProxyRequestBodyTransform, string(changesJSON))
	}
}

func _handleRequestThinkingTransformHeaders(resp *http.Response) {
	thinkingTransformChanges, ok := ctxhelper.GetRequestThinkingTransformChanges(resp.Request.Context())
	if !ok || thinkingTransformChanges == nil {
		return
	}
	v, ok := thinkingTransformChanges.(map[string]any)
	if !ok {
		return
	}
	if changesJSON, err := json.Marshal(v); err == nil {
		resp.Header.Set(vars.XAIProxyRequestThinkingTransform, string(changesJSON))
	}
}

func _handleEnsureContentType(resp *http.Response) {
	if resp.StatusCode == http.StatusOK {
		if ctxhelper.MustGetIsStream(resp.Request.Context()) {
			resp.Header.Set("Content-Type", "text/event-stream")
		}
	}
}

func _handleContentLength(resp *http.Response) {
	// force chunked transfer, worry-free
	resp.Header.Del("Content-Length")
}

func _handleCORS(resp *http.Response) {
	h := resp.Header

	// remove all CORS headers from upstream to avoid duplication
	h.Del(echo.HeaderAccessControlAllowOrigin)
	h.Del(echo.HeaderAccessControlAllowMethods)
	h.Del(echo.HeaderAccessControlAllowHeaders)
	h.Del(echo.HeaderAccessControlAllowCredentials)

	// force enable CORS
	h.Set(echo.HeaderVary, echo.HeaderOrigin)
	h.Set(echo.HeaderAccessControlAllowOrigin, "*")
}
