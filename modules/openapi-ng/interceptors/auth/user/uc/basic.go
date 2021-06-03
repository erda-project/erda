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
	"encoding/base64"
	"encoding/json"
	"mime"
	"net/http"
	"strings"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors/auth"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors/auth/user"
	"github.com/erda-project/erda/pkg/ucauth"
)

// BasicAuther .
type BasicAuther struct {
	name              string
	getUCUserAuth     func() *ucauth.UCUserAuth
	makeSessionCookie func(r *http.Request, token string) (*http.Cookie, error)
	bdl               *bundle.Bundle
}

func (a *BasicAuther) Name() string { return a.name }

func (a *BasicAuther) Match(r *http.Request) (auth.AuthChecker, bool) {
	const prefix = "Basic "
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) <= len(prefix) {
		return nil, false
	}
	if !strings.EqualFold(authHeader[0:len(prefix)], prefix) {
		return nil, false
	}
	authHeader = strings.TrimSpace(authHeader[len(prefix):])
	return func(r *http.Request) (*auth.CheckResult, error) {
		userpwd, err := base64.StdEncoding.DecodeString(authHeader)
		if err != nil {
			return &auth.CheckResult{Success: false}, nil
		}
		splitted := strings.SplitN(string(userpwd), ":", 2)
		if len(splitted) != 2 {
			return &auth.CheckResult{Success: false}, nil
		}
		return a.checkUsernameAndPassword(splitted[0], splitted[1], r)
	}, true
}

func (a *BasicAuther) checkUsernameAndPassword(username, password string, r *http.Request) (*auth.CheckResult, error) {
	// get user info
	info, token, err := a.getUserInfo(username, password)
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
			token: token,
			info:  info,
			orgID: orgID,
		},
	}, nil
}

func (a *BasicAuther) getUserInfo(username, password string) (*ucauth.UserInfo, *ucauth.OAuthToken, error) {
	ucUserAuth := a.getUCUserAuth()
	token, err := ucUserAuth.PwdAuth(username, password)
	if err != nil {
		return nil, nil, err
	}
	info, err := ucUserAuth.GetUserInfo(token)
	if err != nil {
		return nil, nil, err
	}
	info.Token = token.AccessToken
	return &info, &token, nil
}

func (a *BasicAuther) RegisterHandler(add func(method, path string, h http.HandlerFunc)) {
	add(http.MethodPost, "/login", a.Login)
}

func (a *BasicAuther) Login(rw http.ResponseWriter, r *http.Request) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	contentType, _, _ := mime.ParseMediaType(r.Header.Get("content-type"))
	switch contentType {
	case "application/json":
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(rw, "Invalid Request Body", http.StatusBadRequest)
			return
		}
	default:
		request.Username = r.FormValue("username")
		request.Password = r.FormValue("password")
	}

	if request.Username == "" || request.Password == "" {
		http.Error(rw, "username or password is empty", http.StatusBadRequest)
		return
	}

	info, _, err := a.getUserInfo(request.Username, request.Password)
	if err != nil || info == nil {
		http.Error(rw, "Login Failed", http.StatusUnauthorized)
		return
	}

	cookie, err := a.makeSessionCookie(r, info.Token)
	http.SetCookie(rw, cookie)

	byts, err := json.Marshal(struct {
		SessionID string `json:"sessionid"`
		*ucauth.UserInfo
	}{
		SessionID: cookie.Value,
		UserInfo:  info,
	})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(byts)
}
