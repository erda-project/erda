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

package ucauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/pkg/errors"
)

// User 用户中心用户数据结构
type User struct {
	ID        string `json:"user_id"`
	Name      string `json:"username"`
	Nick      string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Phone     string `json:"phone_number"`
	Email     string `json:"email"`
}

// UCClient UC客户端
type UCClient struct {
	baseURL     string
	isOry       bool
	client      *httpclient.HTTPClient
	ucTokenAuth *UCTokenAuth
}

// NewUCClient 初始化UC客户端
func NewUCClient(baseURL, clientID, clientSecret string) *UCClient {
	if clientID == OryCompatibleClientId {
		// TODO: it's a hack
		return &UCClient{
			baseURL: baseURL,
			isOry:   true,
		}
	}
	tokenAuth, err := NewUCTokenAuth(baseURL, clientID, clientSecret)
	if err != nil {
		panic(err)
	}
	return &UCClient{
		baseURL:     baseURL,
		client:      httpclient.New(),
		ucTokenAuth: tokenAuth,
		isOry:       false,
	}
}

func (c *UCClient) oryEnabled() bool {
	return c.isOry
}

func (c *UCClient) oryKratosPrivateAddr() string {
	return c.baseURL
}

// InvalidateServerToken 使 server token 失效
func (c *UCClient) InvalidateServerToken() {
	c.ucTokenAuth.ExpireServerToken()
}

// FindUsers 根据用户ID查找用户信息
func (c *UCClient) FindUsers(ids []string) ([]User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if c.oryEnabled() {
		return getUserByIDs(c.oryKratosPrivateAddr(), ids)
	}
	parts := make([]string, len(ids))
	for _, id := range ids {
		parts = append(parts, "user_id:"+id)
	}
	query := strings.Join(parts, " OR ")

	// 保证返回的用户顺序为 id 列表顺序

	return c.findUsersByQuery(query, ids...)
}

// FindUsersByKey 根据key查找用户，key可匹配用户名/邮箱/手机号
func (c *UCClient) FindUsersByKey(key string) ([]User, error) {
	if key == "" {
		return nil, nil
	}
	if c.oryEnabled() {
		return getUserByKey(c.oryKratosPrivateAddr(), key)
	}
	query := fmt.Sprintf("username:%s OR nickname:%s OR phone_number:%s OR email:%s", key, key, key, key)

	return c.findUsersByQuery(query)
}

// GetUser 获取用户详情
func (c *UCClient) GetUser(userID string) (*User, error) {
	if c.oryEnabled() {
		return getUserByID(c.oryKratosPrivateAddr(), userID)
	}
	var (
		user *User
		err  error
	)
	// 增加重试机制，防止因 uc 升级 serverToken 格式不兼容，无法获取用户信息
	for i := 0; i < 3; i++ {
		user, err = c.getUser(userID)
		if err != nil {
			continue
		}
		return user, nil
	}

	return nil, err
}

func (c *UCClient) getUser(userID string) (*User, error) {
	token, err := c.ucTokenAuth.GetServerToken(false)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get token when finding user")
	}

	var user User
	r, err := c.client.Get(c.baseURL).
		Path(strutil.Concat("/api/open/v1/users/", userID)).
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).
		Do().JSON(&user)
	if err != nil {
		return nil, err
	}

	if !r.IsOK() {
		if r.StatusCode() == http.StatusUnauthorized {
			c.InvalidateServerToken()
		}
		return nil, errors.Errorf("failed to find user, status code: %d", r.StatusCode())
	}
	if user.ID == "" {
		return nil, errors.Errorf("failed to find user %s", userID)
	}

	return &user, nil
}

func (c *UCClient) findUsersByQuery(query string, idOrder ...string) ([]User, error) {
	// 获取token
	token, err := c.ucTokenAuth.GetServerToken(false)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get token when finding users")
	}

	// 批量查询用户
	var users []User
	var b bytes.Buffer
	r, err := c.client.Get(c.baseURL).
		Path("/api/open/v1/users").
		Param("query", query).
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).
		Do().Body(&b)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find users")
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to find users, status code: %d", r.StatusCode())
	}
	content := b.Bytes()
	if err := json.Unmarshal(content, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %v", string(content))
	}

	// 保证顺序
	if len(idOrder) > 0 {
		userMap := make(map[string]User)
		for _, user := range users {
			userMap[user.ID] = user
		}
		var orderedUsers []User
		for _, id := range idOrder {
			orderedUsers = append(orderedUsers, userMap[id])
		}
		return orderedUsers, nil
	}

	return users, nil
}
