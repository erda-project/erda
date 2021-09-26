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

package orykratos

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/modules/core/openapi-ng"
	"github.com/erda-project/erda/modules/core/openapi-ng/common"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/ucauth"
)

const SessionCookieName = "ory_kratos_session"

func (p *provider) addUserInfoAPI(router openapi.Interface) {
	router.Add(http.MethodGet, "/api/users/me", p.GetUserInfo)
	router.Add(http.MethodGet, "/me", p.GetUserInfo)
}

func (p *provider) GetUserInfo(rw http.ResponseWriter, r *http.Request) {
	sessionID := p.getSession(r)
	if len(sessionID) > 0 {
		r = r.WithContext(context.WithValue(r.Context(), "session", sessionID))
	}
	info, err := p.getUserInfo(sessionID)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusUnauthorized)
		return
	}

	common.ResponseJSON(rw, &struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}{
		Success: true,
		Data: map[string]interface{}{
			"id":       info.ID,
			"name":     info.UserName,
			"nick":     info.NickName,
			"avatar":   info.AvatarUrl,
			"phone":    info.Phone,
			"email":    info.Email,
			"token":    info.Token,
			"userType": "new",
		},
	})
}

func (p *provider) getSession(r *http.Request) string {
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil && cookie != nil {
		return cookie.Value
	}
	return ""
}

func (p *provider) getUserInfo(sessionID string) (*ucauth.UserInfo, error) {
	var s OryKratosSession
	r, err := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(p.Cfg.OryKratosAddr).
		Cookie(&http.Cookie{
			Name:  SessionCookieName,
			Value: sessionID,
		}).
		Path("/sessions/whoami").
		Do().JSON(&s)
	if err != nil {
		return nil, err
	}
	if !r.IsOK() {
		return nil, errors.Errorf("get kratos user info error, statusCode: %d", r.StatusCode())
	}

	info := identityToUserInfo(s.Identity)
	ucUserID, err := p.bundle.GetUcUserID(string(info.ID))
	if err != nil {
		return nil, err
	}
	if ucUserID != "" {
		info.ID = ucauth.USERID(ucUserID)
	}
	return info, nil
}

type OryKratosSession struct {
	ID       string            `json:"id"`
	Active   bool              `json:"active"`
	Identity OryKratosIdentity `json:"identity"`
}

type OryKratosIdentity struct {
	ID       ucauth.USERID           `json:"id"`
	SchemaID string                  `json:"schema_id"`
	Traits   OryKratosIdentityTraits `json:"traits"`
}

type OryKratosIdentityTraits struct {
	Email string `json:"email"`
	Name  string `json:"username"`
	Nick  string `json:"nickname"`
	Phone string `json:"phone"`
}

type OryKratosIdentityTraitsName struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

func nameConversion(name OryKratosIdentityTraitsName) string {
	// TODO: eastern name vs western name
	return name.Last + name.First
}

func identityToUser(i OryKratosIdentity) ucauth.User {
	return ucauth.User{
		ID:    string(i.ID),
		Name:  i.Traits.Name,
		Nick:  i.Traits.Nick,
		Email: i.Traits.Email,
		Phone: i.Traits.Phone,
	}
}

func identityToUserInfo(i OryKratosIdentity) *ucauth.UserInfo {
	return userToUserInfo(identityToUser(i))
}

func userToUserInfo(u ucauth.User) *ucauth.UserInfo {
	return &ucauth.UserInfo{
		ID:        ucauth.USERID(u.ID),
		Email:     u.Email,
		Phone:     u.Phone,
		AvatarUrl: u.AvatarURL,
		UserName:  u.Name,
		NickName:  u.Nick,
		Enabled:   true,
	}
}
