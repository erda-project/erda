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

package ucclient

import (
	"context"

	"github.com/pkg/errors"

	useroauthpb "github.com/erda-project/erda-proto-go/core/user/oauth/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/pointer"
)

type UCClient struct {
	baseURL string
	client  *httpclient.HTTPClient
	oauth   useroauthpb.UserOAuthServiceServer
}

func NewUCClient(baseURL string, client *httpclient.HTTPClient, oauth useroauthpb.UserOAuthServiceServer) *UCClient {
	if client == nil {
		client = httpclient.New()
	}
	return &UCClient{
		baseURL: baseURL,
		client:  client,
		oauth:   oauth,
	}
}

func (c *UCClient) ListUCAuditsByLastID(ucAuditReq apistructs.UCAuditsListRequest) (*apistructs.UCAuditsListResponse, error) {
	client, err := c.authedClient(nil)
	if err != nil {
		return nil, err
	}

	var getResp apistructs.UCAuditsListResponse
	resp, err := client.Post(c.baseURL).
		Path("/api/event-log/admin/list-last-event").
		JSONBody(&ucAuditReq).
		Do().JSON(&getResp)
	if err != nil {
		return nil, err
	}
	if !resp.IsOK() {
		return nil, errors.Errorf("failed to list uc audits, status-code: %d", resp.StatusCode())
	}

	return &getResp, nil
}

func (c *UCClient) ListUCAuditsByEventTime(ucAuditReq apistructs.UCAuditsListRequest) (*apistructs.UCAuditsListResponse, error) {
	client, err := c.authedClient(nil)
	if err != nil {
		return nil, err
	}

	var getResp apistructs.UCAuditsListResponse
	resp, err := client.Post(c.baseURL).
		Path("/api/event-log/admin/list-event-time").
		JSONBody(&ucAuditReq).
		Do().JSON(&getResp)
	if err != nil {
		return nil, err
	}
	if !resp.IsOK() {
		return nil, errors.Errorf("failed to list uc audits, status-code: %d", resp.StatusCode())
	}

	return &getResp, nil
}

func (c *UCClient) authedClient(refresh *bool) (*httpclient.HTTPClient, error) {
	if c.oauth == nil {
		return nil, errors.New("oauth service is nil")
	}
	oauthToken, err := c.oauth.ExchangeClientCredentials(
		context.Background(), &useroauthpb.ExchangeClientCredentialsRequest{
			Refresh: pointer.BoolDeref(refresh, false),
		},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to exchange client credentials token")
	}

	return c.client.BearerTokenAuth(oauthToken.AccessToken), nil
}
