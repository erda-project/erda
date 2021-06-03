// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package uc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors/auth"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors/auth/user"
	"github.com/erda-project/erda/pkg/ucauth"
)

// LoginAuther .
type LoginAuther struct {
	name               string
	sessionKey         string
	getUCUserAuth      func() *ucauth.UCUserAuth
	makeSessionCookie  func(r *http.Request, token string) (*http.Cookie, error)
	getSession         func(session string) (string, bool, error)
	bdl                *bundle.Bundle
	publicURL          string
	ucPublicURL        string
	clientID           string
	defaultRedirectURL string
}

func (a *LoginAuther) Name() string { return a.name }
func (a *LoginAuther) Match(r *http.Request) (auth.AuthChecker, bool) {
	session, err := r.Cookie(a.sessionKey)
	if err != nil || len(session.Value) <= 0 {
		return nil, false
	}
	return func(r *http.Request) (*auth.CheckResult, error) {
		// get token from session
		token, ok, err := a.getSession(session.Value)
		if err != nil {
			return nil, err
		}
		if !ok {
			return &auth.CheckResult{Success: false}, nil
		}

		// get user info
		ucUserAuth := a.getUCUserAuth()
		otoken := ucauth.OAuthToken{AccessToken: token}
		info, err := ucUserAuth.GetUserInfo(otoken)
		if err != nil {
			return &auth.CheckResult{Success: false}, nil
		}

		// check org
		ok, orgID, err := user.CheckOrg(a.bdl, r, string(info.ID))
		if err != nil {
			return &auth.CheckResult{Success: false}, nil
		}
		if !ok {
			return &auth.CheckResult{Success: false}, nil
		}

		return &auth.CheckResult{
			Success: true,
			Data: &userInfo{
				token: &otoken,
				info:  &info,
				orgID: orgID,
			},
		}, nil
	}, true
}

func (a *LoginAuther) RegisterHandler(add func(method, path string, h http.HandlerFunc)) {
	add(http.MethodGet, "/api/openapi/login", a.GetLoginURL)
	add(http.MethodGet, "/api/openapi/logincb", a.LoginCallback)
}

func (a *LoginAuther) GetLoginURL(rw http.ResponseWriter, r *http.Request) {
	callbackURL := a.publicURL
	if len(a.publicURL) <= 0 {
		callbackURL = fmt.Sprintf("%s://%s", r.URL.Scheme, r.URL.Host)
	}

	authorizeURL := a.getAuthorizeURL(callbackURL, a.getRedirectURL(r.Header.Get("Referer")))
	body, _ := json.Marshal(struct {
		URL string `json:"url"`
	}{
		URL: authorizeURL,
	})

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(body)
}

func (a *LoginAuther) getAuthorizeURL(callbackURL, referer string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("scope", "public_profile")
	params.Set("client_id", a.clientID)
	redirectURL := fmt.Sprintf("%s/logincb?referer=%s", callbackURL, url.QueryEscape(referer))
	params.Set("redirect_uri", redirectURL)
	return fmt.Sprintf("%s/oauth/authorize?%s", a.ucPublicURL, params.Encode())
}

func (a *LoginAuther) LoginCallback(rw http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	referer := a.getRedirectURL(r.URL.Query().Get("referer"))

	ucUserAuth := a.getUCUserAuth()
	atoken, err := ucUserAuth.Login(code)
	if err != nil {
		http.Error(rw, "Login Failed", http.StatusUnauthorized)
		return
	}
	_, err = ucUserAuth.GetUserInfo(atoken)
	if err != nil {
		http.Error(rw, "Login Failed", http.StatusUnauthorized)
		return
	}
	cookie, err := a.makeSessionCookie(r, atoken.AccessToken)
	if err != nil {
		http.Error(rw, "Login Failed", http.StatusUnauthorized)
		return
	}

	http.SetCookie(rw, cookie)
	http.Redirect(rw, r, referer, http.StatusFound)
}

func (a *LoginAuther) getRedirectURL(referer string) string {
	if len(referer) > 0 {
		return referer
	}
	return a.defaultRedirectURL
}
