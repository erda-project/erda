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

package set_ai_proxy_header

import (
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	custom_http_director "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/custom-http-director"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

const (
	Name = "set-ai-proxy-header"
)

var (
	_ filter_define.ProxyResponseModifier = (*Filter)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type Filter struct {
	filter_define.PassThroughResponseModifier
}

var Creator filter_define.ResponseModifierCreator = func(_ string, _ json.RawMessage) filter_define.ProxyResponseModifier {
	return &Filter{}
}

func (f *Filter) OnHeaders(resp *http.Response) error {
	if model, ok := ctxhelper.GetModel(resp.Request.Context()); ok && model != nil {
		resp.Header.Set(vars.XAIProxyModelId, ctxhelper.MustGetModel(resp.Request.Context()).Id)
		resp.Header.Set(vars.XAIProxyModelName, ctxhelper.MustGetModel(resp.Request.Context()).Name)
		resp.Header.Set(vars.XAIProxyProviderName, ctxhelper.MustGetModelProvider(resp.Request.Context()).Name)
	}

	f.handleRequestIdHeaders(resp)
	f.handleRequestBodyTransformHeaders(resp)

	return nil
}

// handleRequestIdHeaders handles request ID related header settings
func (f *Filter) handleRequestIdHeaders(resp *http.Response) {
	// Handle X-Request-Id returned by LLM backend, rename to X-Request-Id-LLM-Backend
	if backendRequestID := resp.Header.Get(vars.XRequestId); backendRequestID != "" {
		resp.Header.Set(vars.XRequestIdLLMBackend, backendRequestID)
		// Delete original X-Request-Id to avoid conflicts
		resp.Header.Del(vars.XRequestId)
	}

	// Set two independent IDs to response headers
	resp.Header.Set(vars.XRequestId, ctxhelper.MustGetRequestID(resp.Request.Context()))                    // Client-controllable Request ID
	resp.Header.Set(vars.XAIProxyGeneratedCallId, ctxhelper.MustGetGeneratedCallID(resp.Request.Context())) // System-generated Call ID
}

// handleRequestBodyTransformHeaders handles body transformation related headers
func (f *Filter) handleRequestBodyTransformHeaders(resp *http.Response) {
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
