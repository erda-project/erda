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

package iam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/core/user/auth/applier"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/pointer"
)

func (p *provider) Me(_ context.Context, credential *domain.PersistedCredential) (*common.UserInfo, error) {
	switch credential.Authenticator.(type) {
	case *applier.BearerTokenAuth:
		return p.getUserWithOAuthToken(credential)
	case *applier.QueryTokenAuth:
		return p.getUserByAuthToken(credential)
	default:
		return nil, errors.New("not support auth context")
	}
}

func (p *provider) getUserWithOAuthToken(credential *domain.PersistedCredential) (*common.UserInfo, error) {
	var (
		reqPath = "/iam/api/v1/admin/user/me"
		body    bytes.Buffer
	)

	req := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(p.Cfg.BackendHost).
		Path(reqPath)
	credential.Authenticator.Apply(req)

	r, err := req.Do().Body(&body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request iam")
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to call %s, status code: %d, resp body: %s",
			reqPath, r.StatusCode(), body.String())
	}

	var userWithToken common.IAMResponse[common.IAMUserDto]
	if err := json.NewDecoder(&body).Decode(&userWithToken); err != nil {
		return nil, err
	}

	userInfo := userWithToken.Data

	return &common.UserInfo{
		ID:          common.USERID(strconv.FormatInt(userInfo.ID, 10)),
		Email:       pointer.StringDeref(userInfo.Email, ""),
		Phone:       userInfo.Mobile,
		UserName:    userInfo.Username,
		NickName:    userInfo.Nickname,
		LastLoginAt: userInfo.LastLoginAt,
	}, nil
}

func (p *provider) getUserByAuthToken(credential *domain.PersistedCredential) (*common.UserInfo, error) {
	var (
		reqPath = fmt.Sprintf("/%s/iam/api/v1/admin/user/find-by-token", p.Cfg.ApplicationName)
		body    bytes.Buffer
	)

	req := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(p.Cfg.BackendHost).
		Path(reqPath)
	credential.Authenticator.Apply(req)

	r, err := req.Do().Body(&body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request iam")
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to call %s, status code: %d, resp body: %s",
			reqPath, r.StatusCode(), body.String())
	}

	var userWithToken common.IAMResponse[common.IAMUserWithToken]
	if err := json.NewDecoder(&body).Decode(&userWithToken); err != nil {
		return nil, err
	}

	userInfo := userWithToken.Data.User

	var refresh *common.SessionRefresh
	if userWithToken.Data.NewToken != "" {
		refresh = &common.SessionRefresh{
			Token: userWithToken.Data.NewToken,
		}
	}
	if cfg := userWithToken.Data.CookieConfig; cfg != nil {
		if refresh == nil {
			refresh = &common.SessionRefresh{}
		}
		refresh.Cookie = &http.Cookie{
			Name:     cfg.Name,
			Domain:   cfg.Domain,
			Path:     cfg.Path,
			MaxAge:   cfg.MaxAge,
			Secure:   cfg.Secure,
			HttpOnly: cfg.HttpOnly,
		}
	}

	return &common.UserInfo{
		ID:             common.USERID(strconv.FormatInt(userInfo.ID, 10)),
		Email:          pointer.StringDeref(userInfo.Email, ""),
		Phone:          userInfo.Mobile,
		UserName:       userInfo.Username,
		AvatarUrl:      pointer.StringDeref(userInfo.Avatar, ""),
		NickName:       userInfo.Nickname,
		LastLoginAt:    userInfo.LastLoginAt,
		SessionRefresh: refresh,
	}, nil
}
