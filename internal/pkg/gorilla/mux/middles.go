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

package mux

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo"
)

var HTTPMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

var SetXRequestId Middle = func(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := r.Header.Get("X-Request-Id"); id == "" {
			r.Header.Set("X-Request-Id", strings.ReplaceAll(uuid.NewString(), "-", ""))
		}
		h.ServeHTTP(w, r)
	})
}

var CORS Middle = func(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			h.ServeHTTP(w, r)
			w.Header().Set(echo.HeaderVary, echo.HeaderOrigin)
			w.Header().Set(echo.HeaderAccessControlAllowOrigin, "*")
			return
		}

		w.Header().Set(echo.HeaderVary, echo.HeaderOrigin)
		w.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestMethod)
		w.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestHeaders)
		w.Header().Set(echo.HeaderAccessControlAllowOrigin, "*")
		w.Header().Set(echo.HeaderAccessControlAllowMethods, strings.Join(HTTPMethods, ","))
		if v := r.Header.Get(echo.HeaderAccessControlRequestHeaders); v != "" {
			w.Header().Set(echo.HeaderAccessControlAllowHeaders, v)
		}
	})
}
