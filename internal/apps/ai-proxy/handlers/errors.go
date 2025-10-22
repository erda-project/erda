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

package handlers

import (
	"net/http"

	"github.com/pkg/errors"
)

var (
	UserPermissionDenied  = errors.New("user permission denied")
	InvalidSessionId      = errors.New("invalid session id")
	InvalidSessionName    = errors.New("invalid session name")
	InvalidSessionTopic   = errors.New("invalid session topic")
	InvalidSessionSource  = errors.New("invalid session source")
	InvalidSessionModel   = errors.New("invalid session model")
	InvalidSessionResetAt = errors.New("invalid session resetAt")

	ErrAkNotFound        = HTTPError(errors.New("ak not found"), http.StatusUnauthorized)
	ErrNoPermission      = HTTPError(errors.New("no permission"), http.StatusForbidden)
	ErrNoAdminPermission = HTTPError(errors.New("no admin permission"), http.StatusForbidden)
	ErrTokenExpired      = HTTPError(errors.New("token expired, please reapply"), http.StatusForbidden)

	ErrClientIdParamMismatch = HTTPError(errors.New("clientId param mismatch"), http.StatusForbidden)
	ErrTokenIdParamMismatch  = HTTPError(errors.New("tokenId param mismatch"), http.StatusForbidden)
	ErrNotAuthorized         = HTTPError(errors.New("not authorized"), http.StatusUnauthorized)
)

func HTTPError(err error, code int) error {
	if err == nil {
		err = errors.New(http.StatusText(code))
	}
	return httpError{error: err, code: code}
}

type httpError struct {
	error
	code int
}

func (e httpError) HTTPStatus() int {
	return e.code
}
