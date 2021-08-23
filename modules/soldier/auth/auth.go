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

package auth

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/soldier/settings"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type Client struct {
	AccessTokenValiditySeconds  int
	AutoApprove                 bool
	ClientId                    string
	ClientLogoUrl               string
	ClientName                  string
	ClientSecret                string
	RefreshTokenValiditySeconds int
	UserId                      int
}

type ClientToken struct {
	AccessToken string `json:"access_token"`
}

/*
client example
CreateClient{
		AccessTokenValiditySeconds:  433200,
		ClientId:                    clientID,
		ClientName:                  "soldier",
		ClientSecret:                clientSecret,
		RefreshTokenValiditySeconds: 433200,
		UserId: rand.Intn(100)
}
*/

func (c *Client) newClient() error {
	response, err := httpclient.New().Post(settings.OpenAPIURL).
		Path("/api/openapi/manager/clients").JSONBody(c).Do().DiscardBody()
	if err != nil {
		return err
	}
	if !response.IsOK() {
		return fmt.Errorf("get status code %d when create client", response.StatusCode())
	}
	return nil
}

func (c *Client) getToken() (accessToken string, err error) {
	var token ClientToken
	base64Code := base64.StdEncoding.EncodeToString([]byte(c.ClientId + ":" + c.ClientSecret))
	response, err := httpclient.New().Post(settings.OpenAPIURL).
		Path("/api/openapi/client-token").
		Header("Authorization", "Basic "+base64Code).
		Do().JSON(&token)
	if err != nil {
		return "", err
	}
	if !response.IsOK() {
		return "", fmt.Errorf("get status code %d when create client", response.StatusCode())
	}
	return token.AccessToken, nil
}

func (c *Client) GetToken(num int) (string, error) {
	for i := 0; i < num; i++ {
		if err := c.newClient(); err != nil {
			logrus.Error(err)
			continue
		}
		token, err := c.getToken()
		if err != nil {
			logrus.Error(err)
			continue
		}
		return token, nil
	}
	return "", errors.New("reach the retry limit to get token")
}
