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
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/user/identity/pb"
	"github.com/erda-project/erda/internal/core/user/impl/iam"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

func (p *provider) GetCurrentUser(_ context.Context, req *pb.GetCurrentUserRequest) (*pb.GetCurrentUserResponse, error) {
	var (
		user    *commonpb.UserInfo
		refresh *pb.SessionRefresh
		err     error
	)
	switch req.Source {
	case pb.TokenSource_Grant:
		user, err = p.getUserWithGrantedToken(req.AccessToken)
		if err != nil {
			return nil, err
		}
	case pb.TokenSource_Cookie:
		user, refresh, err = p.getUserWithCookie(req.CookieName, req.AccessToken)
		if err != nil {
			return nil, err
		}
	}

	return &pb.GetCurrentUserResponse{
		Data:           user,
		SessionRefresh: refresh,
	}, nil
}

func (p *provider) getUserWithGrantedToken(token string) (*commonpb.UserInfo, error) {
	var (
		reqPath = "/iam/api/v1/admin/user/me"
		body    bytes.Buffer
	)

	req := httpclient.New(httpclient.WithCompleteRedirect()).
		BearerTokenAuth(token).
		Get(p.Cfg.BackendHost).
		Path(reqPath)

	r, err := req.Do().Body(&body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request iam")
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to call %s, status code: %d, resp body: %s",
			reqPath, r.StatusCode(), body.String())
	}

	var userWithToken iam.Response[iam.UserDto]
	if err := json.NewDecoder(&body).Decode(&userWithToken); err != nil {
		return nil, err
	}

	userInfo := userWithToken.Data

	return &commonpb.UserInfo{
		Id:    cast.ToString(userInfo.ID),
		Email: userInfo.Email,
		Phone: userInfo.Mobile,
		Name:  userInfo.Username,
		Nick:  userInfo.Nickname,
	}, nil
}

func (p *provider) getUserWithCookie(name *string, value string) (*commonpb.UserInfo, *pb.SessionRefresh, error) {
	var (
		reqPath = fmt.Sprintf("/%s/iam/api/v1/admin/user/find-by-token", p.Cfg.ApplicationName)
		body    bytes.Buffer
	)

	if name == nil {
		return nil, nil, errors.New("illegal cookie name")
	}
	cookieName := *name

	req := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(p.Cfg.BackendHost).
		Cookie(&http.Cookie{Name: cookieName, Value: value}).
		Path(reqPath)

	r, err := req.Do().Body(&body)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to request iam")
	}
	if !r.IsOK() {
		return nil, nil, fmt.Errorf("failed to call %s, status code: %d, resp body: %s",
			reqPath, r.StatusCode(), body.String())
	}

	var userWithToken iam.Response[iam.UserWithToken]
	if err := json.NewDecoder(&body).Decode(&userWithToken); err != nil {
		return nil, nil, err
	}

	userInfo := userWithToken.Data.User

	var refresh *pb.SessionRefresh
	if userWithToken.Data.NewToken != "" {
		refreshCookie := &pb.CookieRefresh{
			Name:  cookieName,
			Value: userWithToken.Data.NewToken,
			Path:  "/",
		}
		if cfg := userWithToken.Data.CookieConfig; cfg != nil {
			httpOnly := cfg.HttpOnly
			secure := cfg.Secure
			if cfg.Domain != "" {
				refreshCookie.Domain = cfg.Domain
			}
			if cfg.Path != "" {
				refreshCookie.Path = cfg.Path
			}
			refreshCookie.HttpOnly = &httpOnly
			refreshCookie.Secure = &secure
		}
		if userWithToken.Data.Expire > 0 {
			expireAt := time.Now().Add(time.Duration(userWithToken.Data.Expire) * time.Second)
			refreshCookie.ExpireAt = timestamppb.New(expireAt)
			refreshCookie.MaxAge = int32(userWithToken.Data.Expire)
		}
		refresh = &pb.SessionRefresh{
			Cookie: refreshCookie,
		}
	}

	return &commonpb.UserInfo{
		Id:     cast.ToString(userInfo.ID),
		Email:  userInfo.Email,
		Phone:  userInfo.Mobile,
		Name:   userInfo.Username,
		Avatar: userInfo.Avatar,
		Nick:   userInfo.Nickname,
	}, refresh, nil
}
