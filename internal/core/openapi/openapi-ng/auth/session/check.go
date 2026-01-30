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

	openapiauth "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
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
			return true, nil
		}
	}
	return false, nil
}

func (a *loginChecker) Check(r *http.Request, data interface{}, _ openapiauth.Options) (bool, *http.Request, domain.UserAuthState, error) {
	user := a.p.UserAuth.NewState()
	result := user.IsLogin(r)
	if result.Code != domain.AuthSuccess {
		a.p.Log.Debugf("failed to auth: %v", result.Detail)
		return false, r, nil, nil
	}
	return true, r, user, nil
}

type tryLoginChecker struct {
	p *provider
}

func (a *tryLoginChecker) Weight() int64 { return math.MinInt64 }

func (a *tryLoginChecker) Match(r *http.Request, opts openapiauth.Options) (bool, interface{}) {
	check, _ := opts.Get("TryCheckLogin").(bool)
	if check {
		userAuthState := a.p.UserAuth.NewState()
		if userAuthState.IsLogin(r).Code == domain.AuthSuccess {
			return true, ""
		}
	}
	return false, nil
}

func (a *tryLoginChecker) Check(r *http.Request, data interface{}, opts openapiauth.Options) (bool, *http.Request, domain.UserAuthState, error) {
	user := a.p.UserAuth.NewState()
	result := user.IsLogin(r)
	if result.Code == domain.AuthSuccess {
		return true, r, user, nil
	}
	return true, r, nil, nil
}
