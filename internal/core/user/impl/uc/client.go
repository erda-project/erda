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

	"github.com/pkg/errors"
	"github.com/samber/lo"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	useroauthpb "github.com/erda-project/erda-proto-go/core/user/oauth/pb"
	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/pointer"
)

func (p *provider) newAuthedClient(refresh *bool) (*httpclient.HTTPClient, error) {
	oauthToken, err := p.UserOAuthSvc.ExchangeClientCredentials(
		context.Background(), &useroauthpb.ExchangeClientCredentialsRequest{
			Refresh: pointer.BoolDeref(refresh, false),
		},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to exchange client credentials token")
	}

	return p.client.BearerTokenAuth(oauthToken.AccessToken), nil
}

func (p *provider) handleQueryUsers(query string) ([]*commonpb.UserInfo, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	var (
		path = "/api/open/v1/users"
		body bytes.Buffer
	)

	r, err := client.Get(p.Cfg.Host).Path(path).
		Param("query", query).
		Do().Body(&body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query user")
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp []*GetUser
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return nil, err
	}

	return usersMapper(resp), nil
}

func (p *provider) handleGetUser(userID string) (*commonpb.UserInfo, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	var (
		path = fmt.Sprintf("/api/open/v1/users/%s", userID)
		body bytes.Buffer
	)

	r, err := client.Get(p.Cfg.Host).Path(path).
		Do().Body(&body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get user")
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp GetUser
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return nil, err
	}

	return userMapper(&resp), nil
}

func (p *provider) handleListLoginMethod() (*LoginTypes, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	var (
		path = "/api/home/admin/login/style"
		body bytes.Buffer
	)

	r, err := client.Get(p.Cfg.Host).Path(path).
		Do().Body(&body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to freeze user")
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[*LoginTypes]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

func (p *provider) handleGetPwdSecurityConfig() (*PwdSecurityConfig, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	var (
		path = "/api/user/admin/pwd-security-config"
		body bytes.Buffer
	)

	r, err := client.Get(p.Cfg.Host).Path(path).
		Do().Body(&body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get password security config")
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[*PwdSecurityConfig]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

func (p *provider) handleUpdatePwdSecurityConfig(config *PwdSecurityConfig) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	var (
		path = "/api/user/admin/pwd-security-config"
		body bytes.Buffer
	)

	r, err := client.Post(p.Cfg.Host).Path(path).
		JSONBody(config).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to update pwd security config")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp ResponseMeta
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if resp.Success != nil && !*resp.Success {
		return errors.New("failed to update pwd security config")
	}

	return nil
}

func (p *provider) handleFreezeUser(userID, operatorID string) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	var (
		path = "/api/user/admin/freeze/" + userID
		body bytes.Buffer
	)

	r, err := client.Put(p.Cfg.Host).Path(path).
		Param("operatorId", operatorID).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to freeze user")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[bool]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if !resp.Result {
		return errors.New("failed to freeze user")
	}

	return nil
}

func (p *provider) handleUnfreezeUser(userID, operatorID string) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	var (
		path = "/api/user/admin/unfreeze/" + userID
		body bytes.Buffer
	)

	r, err := client.Put(p.Cfg.Host).Path(path).
		Param("operatorId", operatorID).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to unfreeze user")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[bool]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if !resp.Result {
		return errors.New("failed to unfreeze user")
	}

	return nil
}

func (p *provider) handleUpdateLoginMethod(req *UpdateLoginMethodRequest, operatorID string) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	var (
		path = "/api/user/admin/change-full-info"
		body bytes.Buffer
	)

	r, err := client.Post(p.Cfg.Host).Path(path).
		Param("operatorId", operatorID).
		JSONBody(&req).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to invoke change user login method")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[bool]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if !resp.Result {
		return errors.New("failed to invoke change user login method")
	}

	return nil
}

func (p *provider) handleCreateUsers(req *pb.UserCreateRequest, operatorID string) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	users := lo.Map(req.Users, func(item *pb.UserCreateItem, _ int) *CreateUserItem {
		return &CreateUserItem{
			Username: item.Name,
			Nickname: item.Nick,
			Mobile:   item.Phone,
			Email:    item.Email,
			Password: item.Password,
		}
	})

	var (
		path = "/api/user/admin/batch-create-user"
		body bytes.Buffer
	)

	r, err := client.Post(p.Cfg.Host).Path(path).
		Param("operatorId", operatorID).
		JSONBody(&CreateUserRequest{Users: users}).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to invoke create user")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[bool]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if !resp.Result {
		return errors.New("failed to invoke create user")
	}
	return nil
}

func (p *provider) handleUpdateUserInfo(req *UpdateUserInfoRequest, operatorID string) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	var (
		path = "/api/user/admin/change-full-info"
		body bytes.Buffer
	)

	r, err := client.Post(p.Cfg.Host).Path(path).
		Param("operatorId", operatorID).
		JSONBody(&req).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to invoke update user info")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}
	return nil
}
