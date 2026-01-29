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

package auth

import (
	"net/http"
	"strconv"

	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/common"
)

func ApplyUserInfoHeaders(req *http.Request, user domain.UserAuthState) domain.UserAuthResult {
	userinfo, r := user.GetInfo(req)
	if r.Code != domain.AuthSuccess {
		return r
	}
	// set User-ID
	req.Header.Set("User-ID", userinfo.Id)

	// with session refresh context
	if newCtx := WithSessionRefresh(req.Context(), userinfo.SessionRefresh); newCtx != req.Context() {
		*req = *req.WithContext(newCtx)
	}

	var scopeinfo common.UserScopeInfo
	scopeinfo, r = user.GetScopeInfo(req)
	if r.Code != domain.AuthSuccess {
		return r
	}
	// set Org-ID
	if scopeinfo.OrgID != 0 {
		req.Header.Set("Org-ID", strconv.FormatUint(scopeinfo.OrgID, 10))
	}
	return domain.UserAuthResult{Code: domain.AuthSuccess}
}
