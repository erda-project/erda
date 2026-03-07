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
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

type modelRetryMetaHeader struct {
	RawLLMBackendRequestCount int    `json:"raw_llm_backend_request_count"`
	RawLLMBackendRetryCount   int    `json:"raw_llm_backend_retry_count"`
	FinalModelInstanceID      string `json:"final_model_instance_id,omitempty"`
}

func noteRetryAuditMetadata(ctx context.Context) {
	rawLLMBackendRequestCount, ok := ctxhelper.GetModelRetryRawLLMBackendRequestCount(ctx)
	if !ok || rawLLMBackendRequestCount <= 1 {
		return
	}
	audithelper.Note(ctx, "model_retry.raw_llm_backend_request_count", rawLLMBackendRequestCount)
	audithelper.Note(ctx, "model_retry.raw_llm_backend_retry_count", rawLLMBackendRequestCount-1)
}

func _handleModelRetryMetaHeader(resp *http.Response) {
	if enabled, ok := ctxhelper.GetModelRetryResponseHeaderMetaEnabled(resp.Request.Context()); ok && !enabled {
		return
	}
	rawLLMBackendRequestCount, ok := ctxhelper.GetModelRetryRawLLMBackendRequestCount(resp.Request.Context())
	if !ok || rawLLMBackendRequestCount <= 1 {
		return
	}
	payload := modelRetryMetaHeader{
		RawLLMBackendRequestCount: rawLLMBackendRequestCount,
		RawLLMBackendRetryCount:   rawLLMBackendRequestCount - 1,
	}
	if model, ok := ctxhelper.GetModel(resp.Request.Context()); ok && model != nil && model.Id != "" {
		payload.FinalModelInstanceID = model.Id
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	resp.Header.Set(vars.XAIProxyModelRetryMeta, string(b))
}
