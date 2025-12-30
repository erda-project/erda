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

package utils

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/sirupsen/logrus"
)

const (
	loginPath      = "/nacos/v1/auth/login"
	namespacesPath = "/nacos/v1/console/namespaces"
	configsPath    = "/nacos/v1/cs/configs"
)

func NewNacosClient(clusterName, addr, user, password string) *NacosClient {
	return &NacosClient{
		addr:        addr,
		user:        user,
		password:    password,
		clusterName: clusterName,
	}
}

type NacosClient struct {
	clusterName string
	addr        string
	user        string
	password    string
	// bearerToken keeps the token without the "Bearer " prefix.
	bearerToken string
}

func (c *NacosClient) Login() (string, error) {
	form := url.Values{}
	form.Set("username", c.user)
	form.Set("password", c.password)
	var result map[string]interface{}
	resp, err := c.client().
		Post(c.addr).Path(loginPath).FormBody(form).Do().JSON(&result)
	if err != nil {
		return "", err
	}
	if !resp.IsOK() {
		return "", fmt.Errorf("nacos login response code error[%d], body:%s", resp.StatusCode(), string(resp.Body()))
	}

	token, err := c.extractToken(result)
	if err != nil {
		return "", err
	}

	c.bearerToken = token

	return c.bearerToken, nil
}

func (c *NacosClient) GetNamespaceId(namespaceName string) (string, error) {
	hc := c.tryGetAuthedClient()

	path := namespacesPath
	var result struct {
		Data []struct {
			NamespaceShowName string
			Namespace         string
		}
	}
	resp, err := hc.Get(c.addr).Path(path).Do().JSON(&result)
	if err != nil {
		return "", err
	}
	if !resp.IsOK() {
		return "", fmt.Errorf("nacos get namespaceid response code error[%d], body:%s", resp.StatusCode(), string(resp.Body()))
	}

	for _, namespace := range result.Data {
		if namespace.NamespaceShowName == namespaceName {
			return namespace.Namespace, nil
		}
	}

	return "", nil
}

func (c *NacosClient) CreateNamespace(namespaceName string) (string, error) {
	hc := c.tryGetAuthedClient()

	params := url.Values{}
	params.Set("namespaceName", namespaceName)
	params.Set("namespaceDesc", namespaceName)
	params.Set("customNamespaceId", namespaceName)
	resp, err := hc.Post(c.addr).Path(namespacesPath).Params(params).Do().DiscardBody()
	if err != nil {
		return "", err
	}
	if !resp.IsOK() {
		return "", fmt.Errorf("nacos create namespace response code error[%d], body:%s", resp.StatusCode(), string(resp.Body()))
	}

	return c.GetNamespaceId(namespaceName)
}

func (c *NacosClient) SaveConfig(tenantName string, groupName string, dataId string, content string) error {
	hc := c.tryGetAuthedClient()

	params := url.Values{}
	params.Set("dataId", dataId)
	params.Set("group", groupName)
	params.Set("content", content)
	params.Set("tenant", tenantName)
	resp, err := hc.Post(c.addr).Path(configsPath).Params(params).Do().DiscardBody()
	if err != nil {
		return err
	}
	if !resp.IsOK() {
		return fmt.Errorf("nacos save config response code error[%d], body:%s", resp.StatusCode(), string(resp.Body()))
	}
	return nil
}

func (c *NacosClient) DeleteConfig(tenantName string, groupName string) error {
	hc := c.tryGetAuthedClient()

	params := url.Values{}
	params.Set("dataId", "application.yml")
	params.Set("group", groupName)
	params.Set("tenant", tenantName)
	resp, err := hc.Delete(c.addr).Path(configsPath).Params(params).Do().DiscardBody()
	if err != nil {
		return err
	}
	if !resp.IsOK() {
		return fmt.Errorf("nacos delete config response code error[%d], body:%s", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

func (c *NacosClient) tryGetAuthedClient() *httpclient.HTTPClient {
	if _, err := c.ensureToken(); err != nil {
		logrus.Warnf("Nacos failed authentication; authentication may not be enabled. error: %v", err)
		return c.client()
	}
	return c.client().BearerTokenAuth(c.bearerToken)
}

func (c *NacosClient) ensureToken() (string, error) {
	if c.bearerToken != "" {
		return c.bearerToken, nil
	}
	return c.Login()
}

func (c *NacosClient) extractToken(result map[string]interface{}) (string, error) {
	// Prefer explicit accessToken, fall back to data for backwards compatibility.
	rawToken, ok := result["accessToken"]
	if !ok {
		rawToken = result["data"]
	}
	if rawToken == nil {
		return "", fmt.Errorf("nacos login success but token missing")
	}

	token := strings.TrimSpace(fmt.Sprint(rawToken))
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[len("bearer "):])
	}
	if token == "" {
		return "", fmt.Errorf("nacos login success but token missing")
	}
	return token, nil
}

func (c *NacosClient) client() *httpclient.HTTPClient {
	return httpclient.New(httpclient.WithClusterDialer(c.clusterName))
}
