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

package http_error

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
)

type HTTPError struct {
	Ctx         context.Context `json:"-"`
	StatusCode  int             `json:"-"`
	Message     string          `json:"message"`
	ErrorCtx    map[string]any  `json:"error,omitempty"`
	AIProxyMeta map[string]any  `json:"ai_proxy_meta,omitempty"`
}

func (he *HTTPError) Error() string {
	errB, _ := json.Marshal(&he.ErrorCtx)
	return fmt.Sprintf("HTTPError: StatusCode: %d (%s), Message: %s, Error: %v", he.StatusCode, http.StatusText(he.StatusCode), he.Message, string(errB))
}

func (he *HTTPError) WriteJSONHTTPError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(he.StatusCode)
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(&he)
	_, _ = w.Write(body.Bytes())
	audithelper.Note(he.Ctx, "status", he.StatusCode)
	audithelper.Note(he.Ctx, "response_body", body.String())
}

func NewHTTPError(ctx context.Context, statusCode int, message string) *HTTPError {
	return &HTTPError{
		Ctx:        ctx,
		StatusCode: statusCode,
		Message:    message,
	}
}

func NewHTTPErrorWithCtx(ctx context.Context, statusCode int, message string, errCtx map[string]any) *HTTPError {
	return &HTTPError{
		Ctx:        ctx,
		StatusCode: statusCode,
		Message:    message,
		ErrorCtx:   errCtx,
	}
}
