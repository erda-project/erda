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

package error_status_handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	httperror "github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

type Filter struct {
	statusCode int
	isError    bool
	bodyBuffer []byte
}

var (
	_ filter_define.ProxyResponseModifier = (*Filter)(nil)
)

var ResponseModifierCreator filter_define.ResponseModifierCreator = func(name string, _ json.RawMessage) filter_define.ProxyResponseModifier {
	return &Filter{}
}

func init() {
	filter_define.RegisterFilterCreator("error-status-handler", ResponseModifierCreator)
}

func (f *Filter) OnHeaders(resp *http.Response) error {
	// Only record status code in OnHeaders phase
	if resp.StatusCode >= 400 {
		f.statusCode = resp.StatusCode
		f.isError = true
		f.bodyBuffer = nil // Reset buffer
	}
	return nil
}

func (f *Filter) OnBodyChunk(resp *http.Response, chunk []byte, index int64) ([]byte, error) {
	if f.isError {
		// If error response, collect all response body data but don't pass to downstream
		f.bodyBuffer = append(f.bodyBuffer, chunk...)
		return nil, nil // Swallow error response chunks
	}
	return chunk, nil // Normal response passes through directly
}

func (f *Filter) OnComplete(resp *http.Response) ([]byte, error) {
	if f.isError {
		// Generate standardized error response in OnComplete phase
		errCtx := map[string]any{
			"type": "llm-backend-error",
		}

		// Try to parse backend response as JSON, if so preserve it
		var jsonResp interface{}
		if json.Unmarshal(f.bodyBuffer, &jsonResp) == nil {
			errCtx["raw_llm_backend_response"] = jsonResp
		} else {
			errCtx["raw_llm_backend_response"] = string(f.bodyBuffer)
		}

		httpErr := httperror.NewHTTPErrorWithCtx(
			f.statusCode,
			"LLM Backend Error",
			errCtx,
		)
		httpErr.AIProxyMeta = map[string]any{
			vars.XAIProxyModel:           ctxhelper.MustGetModel(resp.Request.Context()).Name,
			vars.XRequestId:              ctxhelper.MustGetRequestID(resp.Request.Context()),
			vars.XAIProxyGeneratedCallId: ctxhelper.MustGetGeneratedCallID(resp.Request.Context()),
		}

		errBody, err := json.MarshalIndent(httpErr, "", "  ")
		if err != nil {
			errBody = []byte(fmt.Sprintf(`{"message": %#v}`, httpErr))
		}

		// Modify response status code and headers
		resp.Status = http.StatusText(f.statusCode)
		resp.Header.Set("Content-Type", "application/json")
		resp.Header.Del("Content-Length") // Delete original length, let system recalculate

		return errBody, nil
	}
	return nil, nil
}
