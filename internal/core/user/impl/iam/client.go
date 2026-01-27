// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/pkg/errors"
	"k8s.io/utils/pointer"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	useroauthpb "github.com/erda-project/erda-proto-go/core/user/oauth/pb"
	"github.com/erda-project/erda/pkg/http/httpclient"
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

func (p *provider) getUser(userID string, plainText bool) (*commonpb.UserInfo, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	var (
		path = fmt.Sprintf("/iam/api/v1/admin/user/%s/find", userID)
		resp Response[UserDto]
	)

	if plainText {
		path = fmt.Sprintf("/iam/api/v1/admin/user/%s/plaintext/find", userID)
	}

	r, err := client.Get(p.Cfg.Host).
		Path(path).
		Do().JSON(&resp)
	if err != nil {
		return nil, err
	}

	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d", path, r.StatusCode())
	}
	if resp.Data.ID == 0 {
		return nil, errors.Errorf("failed to find user %s", userID)
	}

	return userMapper(&resp.Data), nil
}

func (p *provider) updateProfile(userId string, newVal map[string]any) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	var (
		path = "/iam/api/v1/admin/user/%s/update"
		body bytes.Buffer
	)

	r, err := client.Post(p.Cfg.Host).Path(fmt.Sprintf(path, userId)).
		JSONBody(newVal).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to update user profile")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, body: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[bool]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if !resp.Data {
		return errors.New("failed to update user profile")
	}

	return nil
}

func (p *provider) userUnlock(userId string) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	var (
		path = "/iam/api/v1/admin/user/%s/unlock"
		body bytes.Buffer
	)

	r, err := client.Post(p.Cfg.Host).Path(fmt.Sprintf(path, userId)).
		JSONBody(map[string]any{}).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to create user")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, body: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[bool]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if !resp.Data {
		return errors.New("failed to unlock user")
	}

	return nil
}

func (p *provider) pagingQuery(no, size int64, conditions map[string]any, plainText bool) (*PagingData[[]*UserDto], error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	conditions["no"] = no
	conditions["size"] = size

	var (
		path = "/iam/api/v1/admin/user/paging"
		body bytes.Buffer
	)

	if plainText {
		path = "/iam/api/v1/admin/user/plaintext/paging"
	}

	r, err := client.Post(p.Cfg.Host).Path(path).
		JSONBody(conditions).
		Do().Body(&body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find users by query")
	}

	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d, body: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[*PagingData[[]*UserDto]]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (p *provider) findUsersByQuery(fieldName, key string) ([]*UserDto, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	var (
		path = "/iam/api/v1/admin/user/list"
		resp Response[[]*UserDto]
	)

	r, err := client.Post(p.Cfg.Host).Path(path).
		JSONBody(map[string]string{
			fieldName: key,
		}).
		Do().JSON(&resp)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find users by query")
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d", path, r.StatusCode())
	}

	return resp.Data, nil
}

func (p *provider) findByIDs(ids []int, plainText bool) ([]*commonpb.UserInfo, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	var (
		path = "/iam/api/v1/admin/user/find-by-ids"
		body bytes.Buffer
	)

	if plainText {
		path = "/iam/api/v1/admin/user/plaintext/find-by-ids"
	}

	r, err := client.Post(p.Cfg.Host).
		Path(path).JSONBody(map[string][]int{
		"userIds": ids,
	}).Do().Body(&body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find users by ids")
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[[]UserDto]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return nil, err
	}

	userList := make([]*commonpb.UserInfo, len(resp.Data))
	for i, user := range resp.Data {
		userList[i] = userMapper(&user)
	}

	return userList, nil
}
