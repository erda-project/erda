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
	"errors"
	"net/http"

	"github.com/erda-project/erda-proto-go/core/user/oauth/pb"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/common"
)

func (p *provider) LoginURL(rw http.ResponseWriter, r *http.Request) {
	referer := r.Header.Get("Referer")
	if len(referer) <= 0 {
		referer = p.Cfg.RedirectAfterLogin
	}

	authURL, err := p.UserOauthSessionSvc.AuthURL(r.Context(), &pb.AuthURLRequest{
		Referer: referer,
	})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	common.ResponseJSON(rw, &struct {
		URL string `json:"url"`
	}{
		URL: authURL.Data,
	})
}

func (p *provider) LoginCallback(rw http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	code := queryValues.Get("code")
	referer := queryValues.Get("referer")
	state := queryValues.Get("state")

	redirectAfterLogin := state
	if redirectAfterLogin == "" {
		redirectAfterLogin = referer
	}

	if redirectAfterLogin == "" {
		err := errors.New("missing redirect url after login")
		p.Log.Error(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if code == "" {
		err := errors.New("missing oauth code")
		p.Log.Error(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	user := p.UserAuth.NewState()
	if err := user.Login(code, queryValues); err != nil {
		p.Log.Errorf("failed to login: %v", err)
		http.Error(rw, err.Error(), http.StatusUnauthorized)
		return
	}

	if !p.referMatcher.Match(redirectAfterLogin) {
		http.Error(rw, "invalid referer", http.StatusBadRequest)
		return
	}

	http.Redirect(rw, r, redirectAfterLogin, http.StatusFound)
}

func (p *provider) Logout(rw http.ResponseWriter, r *http.Request) {
	referer := r.Header.Get("Referer")
	if len(referer) <= 0 {
		referer = p.Cfg.RedirectAfterLogin
	}

	logoutURL, err := p.UserOauthSessionSvc.LogoutURL(r.Context(), &pb.LogoutURLRequest{
		Referer: referer,
	})
	if err != nil {
		p.Log.Errorf("failed to get logout url, %v", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	common.ResponseJSON(rw, &struct {
		URL string `json:"url"`
	}{
		URL: logoutURL.Data,
	})
}
