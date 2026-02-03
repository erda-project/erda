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

package oauth

import (
	"math"
	"net/http"
	"strconv"

	openapiauth "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/auth/sessionrefresh"
	"github.com/erda-project/erda/internal/core/user/common"
)

type loginChecker struct {
	p *provider
}

func (a *loginChecker) Weight() int64 { return a.p.Cfg.Weight }

func (a *loginChecker) Match(r *http.Request, opts openapiauth.Options) (bool, interface{}) {
	check, _ := opts.Get("CheckLogin").(bool)
	if check {
		userAuthState := a.p.UserAuth.NewState()
		if userAuthState.IsLogin(r).Code == domain.AuthSuccess {
			return true, userAuthState
		}
	}
	return false, nil
}

func (a *loginChecker) Check(r *http.Request, data interface{}, _ openapiauth.Options) (bool, *http.Request, error) {
	user := data.(domain.UserAuthState)
	result := applyUserInfoHeaders(r, user)
	if result.Code != domain.AuthSuccess {
		a.p.Log.Debugf("failed to auth: %v", result.Detail)
		return false, r, nil
	}
	return true, r, nil
}

type tryLoginChecker struct {
	p *provider
}

func (a *tryLoginChecker) Weight() int64 { return math.MinInt64 }

func (a *tryLoginChecker) Match(r *http.Request, opts openapiauth.Options) (bool, interface{}) {
	check, _ := opts.Get("TryCheckLogin").(bool)
	if check {
		return true, ""
	}
	return false, nil
}

func (a *tryLoginChecker) Check(r *http.Request, data interface{}, opts openapiauth.Options) (bool, *http.Request, error) {
	user := a.p.UserAuth.NewState()
	result := user.IsLogin(r)
	if result.Code == domain.AuthSuccess {
		applyUserInfoHeaders(r, user)
	}
	return true, r, nil
}

func applyUserInfoHeaders(req *http.Request, user domain.UserAuthState) domain.UserAuthResult {
	userinfo, r := user.GetInfo(req)
	if r.Code != domain.AuthSuccess {
		return r
	}
	// set User-ID
	req.Header.Set("User-ID", userinfo.Id)

	// with session refresh context
	if newCtx := sessionrefresh.With(req.Context(), userinfo.SessionRefresh); newCtx != req.Context() {
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
