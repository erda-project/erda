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

	"github.com/pkg/errors"
	"github.com/spf13/cast"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/user/identity/pb"
	"github.com/erda-project/erda/internal/core/user/impl/uc"
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
		user, refresh, err = p.getUserWithCookie(req.AccessToken)
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
		reqPath = "/api/oauth/me"
		body    bytes.Buffer
	)

	req := httpclient.New(httpclient.WithCompleteRedirect()).
		BearerTokenAuth(token).
		Get(p.Cfg.BackendHost).
		Path(reqPath)

	r, err := req.Do().Body(&body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request uc")
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to call %s, status code: %d, resp body: %s",
			reqPath, r.StatusCode(), body.String())
	}

	var info uc.UserDto
	if err := json.NewDecoder(&body).Decode(&info); err != nil {
		return nil, err
	}

	return &commonpb.UserInfo{
		Id:     cast.ToString(info.Id),
		Name:   info.Username,
		Nick:   info.Nickname,
		Avatar: info.Avatar,
		Phone:  info.Mobile,
		Email:  info.Mobile,
	}, nil
}

func (p *provider) getUserWithCookie(value string) (*commonpb.UserInfo, *pb.SessionRefresh, error) {
	var (
		reqPath = "/api/user/web/current-user"
		body    bytes.Buffer
	)

	req := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(p.Cfg.BackendHost).
		Header("Cookie", fmt.Sprintf("%s=%s", p.Cfg.CookieName, value)).
		Path(reqPath)

	r, err := req.Do().Body(&body)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to request uc")
	}
	if !r.IsOK() {
		return nil, nil, fmt.Errorf("failed to call %s, status code: %d, resp body: %s",
			reqPath, r.StatusCode(), body.String())
	}

	var resp uc.Response[uc.UserDto]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return nil, nil, err
	}

	setCookie := r.ResponseHeader("Set-Cookie")
	cookie, err := http.ParseSetCookie(setCookie)
	if err != nil {
		return nil, nil, err
	}

	sessionRefresh := &pb.SessionRefresh{
		NewToken: cookie.Value,
		ExpireAt: cookie.Expires.Unix(),
	}

	user := resp.Result
	return &commonpb.UserInfo{
		Id:     cast.ToString(user.Id),
		Name:   user.Username,
		Nick:   user.Nickname,
		Avatar: user.Avatar,
		Phone:  user.Mobile,
		Email:  user.Mobile,
	}, sessionRefresh, nil
}

func (p *provider) WriteRefresh(rw http.ResponseWriter, req *http.Request, refresh *pb.SessionRefresh) error {
	return nil
}
