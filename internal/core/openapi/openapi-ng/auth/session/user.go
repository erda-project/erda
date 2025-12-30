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
	"context"
	"net/http"

	"github.com/erda-project/erda/internal/core/openapi/openapi-ng"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/common"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
)

func (p *provider) addUserInfoAPI(router openapi.Interface) {
	router.Add(http.MethodGet, "/api/users/me", p.GetUserInfo)
	router.Add(http.MethodGet, "/me", p.GetUserInfo)
}

func (p *provider) GetUserInfo(rw http.ResponseWriter, r *http.Request) {
	_, err := p.CredStore.Load(context.Background(), r)
	if err != nil {
		http.Error(rw, "lack of required auth credential", http.StatusUnauthorized)
		return
	}

	user := p.UserAuth.NewUserState()
	info, authr := user.GetInfo(r)
	if authr.Code != domain.AuthSuccess {
		http.Error(rw, authr.Detail, authr.Code)
		return
	}
	common.ResponseJSON(rw, &struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}{
		Success: true,
		Data: map[string]interface{}{
			"id":          info.ID,
			"name":        info.UserName,
			"nick":        info.NickName,
			"avatar":      info.AvatarUrl,
			"phone":       info.Phone,
			"email":       info.Email,
			"token":       info.Token,
			"lastLoginAt": info.LastLoginAt,
		},
	})
}
