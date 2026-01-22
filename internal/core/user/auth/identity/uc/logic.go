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

package uc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/core/user/auth/applier"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

func (p *provider) Me(_ context.Context, credential *domain.PersistedCredential) (*common.UserInfo, error) {
	switch credential.Authenticator.(type) {
	case *applier.BearerTokenAuth:
		return p.getUserWithOAuthToken(credential)
	case *applier.CookieTokenAuth:
		return p.getUserWithCookie(credential)
	default:
		return nil, errors.New("not support auth context")
	}
}

func (p *provider) getUserWithOAuthToken(credential *domain.PersistedCredential) (*common.UserInfo, error) {
	var (
		reqPath = "/api/oauth/me"
		body    bytes.Buffer
	)

	req := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(p.Cfg.BackendHost).
		Path(reqPath)
	credential.Authenticator.Apply(req)

	r, err := req.Do().Body(&body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request uc")
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to call %s, status code: %d, resp body: %s",
			reqPath, r.StatusCode(), body.String())
	}

	var info common.UserInfo
	if err := json.NewDecoder(&body).Decode(&info); err != nil {
		return nil, err
	}

	info.SessionRefresh = &common.SessionRefresh{
		Token:     credential.AccessToken,
		SessionID: credential.SessionID,
	}
	return &info, nil
}

func (p *provider) getUserWithCookie(credential *domain.PersistedCredential) (*common.UserInfo, error) {
	var (
		reqPath = "/api/user/web/current-user"
		body    bytes.Buffer
	)

	req := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(p.Cfg.BackendHost).
		Path(reqPath)
	credential.Authenticator.Apply(req)

	r, err := req.Do().Body(&body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request uc")
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to call %s, status code: %d, resp body: %s",
			reqPath, r.StatusCode(), body.String())
	}

	var info common.UCResponse[common.UCCurrentUser]
	if err := json.NewDecoder(&body).Decode(&info); err != nil {
		return nil, err
	}
	if info.Result.ID == "" {
		return nil, errors.New("not login")
	}

	lastLogin := ""
	if info.Result.LastLoginAt > 0 {
		lastLogin = time.Unix(int64(info.Result.LastLoginAt/1e3), 0).Format("2006-01-02 15:04:05")
	}

	setCookie := r.ResponseHeader("Set-Cookie")
	cookie, err := http.ParseSetCookie(setCookie)
	if err != nil {
		return nil, err
	}

	refresh := &common.SessionRefresh{
		Cookie: cookie,
	}

	return &common.UserInfo{
		ID:             info.Result.ID,
		Email:          info.Result.Email,
		Phone:          info.Result.Mobile,
		UserName:       info.Result.Username,
		NickName:       info.Result.Nickname,
		LastLoginAt:    lastLogin,
		SessionRefresh: refresh,
	}, nil
}
