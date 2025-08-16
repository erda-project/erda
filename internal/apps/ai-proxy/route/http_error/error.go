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
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPError struct {
	StatusCode  int            `json:"-"`
	Message     string         `json:"message"`
	ErrorCtx    map[string]any `json:"error,omitempty"`
	AIProxyMeta map[string]any `json:"ai_proxy_meta,omitempty"`
}

func (he *HTTPError) Error() string {
	errB, _ := json.Marshal(&he.ErrorCtx)
	return fmt.Sprintf("HTTPError: StatusCode: %d (%s), Message: %s, Error: %v", he.StatusCode, http.StatusText(he.StatusCode), he.Message, string(errB))
}

func (he *HTTPError) WriteJSONHTTPError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(he.StatusCode)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(&he)
}

func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func NewHTTPErrorWithCtx(statusCode int, message string, errCtx map[string]any) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
		ErrorCtx:   errCtx,
	}
}
