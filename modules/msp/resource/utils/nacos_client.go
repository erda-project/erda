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

package utils

import (
	"fmt"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"net/url"
)

func NewNacosClient(addr, user, password string) *NacosClient {
	return &NacosClient{
		addr:     addr,
		user:     user,
		password: password,
	}
}

type NacosClient struct {
	addr        string
	user        string
	password    string
	accessToken string
}

func (c *NacosClient) Login() (string, error) {
	loginUrl := "/nacos/v1/auth/login?username=" + c.user + "&password=" + c.password
	var result map[string]interface{}
	resp, err := httpclient.New().Post(c.addr).Path(loginUrl).Do().JSON(&result)
	if err != nil {
		return "", err
	}
	if !resp.IsOK() {
		return "", fmt.Errorf("response code error")
	}
	accessToken, ok := result["accessToken"]
	if !ok {
		accessToken = result["data"]
	}

	c.accessToken = fmt.Sprint(accessToken)

	return c.accessToken, nil
}

func (c *NacosClient) GetNamespaceId(namespaceName string) (string, error) {

	if len(c.accessToken) == 0 {
		c.Login()
	}

	path := "/nacos/v1/console/namespaces"
	var result struct {
		Data []struct {
			NamespaceShowName string
			Namespace         string
		}
	}
	resp, err := httpclient.New().TokenAuth(c.accessToken).Get(c.addr).Path(path).Do().JSON(&result)
	if err != nil {
		return "", err
	}
	if !resp.IsOK() {
		return "", fmt.Errorf("response code error")
	}

	for _, namespace := range result.Data {
		if namespace.NamespaceShowName == namespaceName {
			return namespace.Namespace, nil
		}
	}

	return "", nil
}

func (c *NacosClient) CreateNamespace(namespaceName string) (string, error) {

	if len(c.accessToken) == 0 {
		c.Login()
	}

	path := "/nacos/v1/console/namespaces" + "?namespaceName=" + namespaceName + "&namespaceDesc=" + namespaceName + "&customNamespaceId=" + namespaceName
	resp, err := httpclient.New().TokenAuth(c.accessToken).Post(c.addr).Path(path).Do().DiscardBody()
	if err != nil {
		return "", err
	}
	if !resp.IsOK() {
		return "", fmt.Errorf("response code error")
	}

	return c.GetNamespaceId(namespaceName)
}

func (c *NacosClient) SaveConfig(tenantName string, groupName string, dataId string, content string) error {
	if len(c.accessToken) == 0 {
		c.Login()
	}

	path := "/nacos/v1/cs/configs?dataId=" + dataId + "&group=" + groupName + "&content=" + url.QueryEscape(content) + "&tenant=" + tenantName
	resp, err := httpclient.New().TokenAuth(c.accessToken).Post(c.addr).Path(path).Do().DiscardBody()
	if err != nil {
		return err
	}
	if !resp.IsOK() {
		return fmt.Errorf("response code error")
	}
	return nil
}

func (c *NacosClient) DeleteConfig(tenantName string, groupName string) error {
	if len(c.accessToken) == 0 {
		c.Login()
	}

	path := "/nacos/v1/cs/configs?dataId=application.yml&group=" + groupName + "&tenant=" + tenantName
	resp, err := httpclient.New().TokenAuth(c.accessToken).Delete(c.addr).Path(path).Do().DiscardBody()
	if err != nil {
		return err
	}
	if !resp.IsOK() {
		return fmt.Errorf("response code error")
	}

	return nil
}
