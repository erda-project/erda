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
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	httperror "github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
)

var MyErrorHandler = func() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		// sink audit when error
		defer func() {
			audithelper.NoteOnce(r.Context(), "response_chunk_done_at", time.Now())
			audithelper.Flush(r.Context())
		}()

		var defaultStatus = http.StatusBadRequest
		// check error at request rewrite stage
		if rewriteErr, _ := ctxhelper.GetReverseProxyRequestRewriteError(r.Context()); rewriteErr != nil {
			audithelper.Note(r.Context(), "request.rewrite.error.filter", rewriteErr.FilterName)
			audithelper.Note(r.Context(), "request.rewrite.error.message", rewriteErr.Error.Error())
			if rewriteErr.FilterName != "" {
				err = fmt.Errorf("%s: %w", rewriteErr.FilterName, rewriteErr.Error)
			} else {
				err = rewriteErr.Error
			}
		}

		// check error at response modify stage
		if responseErr, _ := ctxhelper.GetReverseProxyResponseModifyError(r.Context()); responseErr != nil {
			audithelper.Note(r.Context(), "response.modify.error.filter", responseErr.FilterName)
			audithelper.Note(r.Context(), "response.modify.error.message", responseErr.Error.Error())
			if responseErr.FilterName != "" {
				err = fmt.Errorf("%s: %w", responseErr.FilterName, responseErr.Error)
			} else {
				err = responseErr.Error
			}
			defaultStatus = http.StatusInternalServerError
		}

		// set ai-proxy response header
		handleAIProxyResponseHeader(&http.Response{
			Header:     w.Header(),
			Request:    r,
			StatusCode: defaultStatus,
			Status:     http.StatusText(defaultStatus),
		})

		var httpError *httperror.HTTPError
		switch {
		case errors.As(err, &httpError):
			httpError.WriteJSONHTTPError(w)
			return
		default:
			httperror.NewHTTPError(r.Context(), defaultStatus, err.Error()).WriteJSONHTTPError(w)
			return
		}
	}
}
