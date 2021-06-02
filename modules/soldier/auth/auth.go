// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
