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

package ucauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
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

type UcUser struct {
	ID        int    `json:"user_id"`
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

func (c *UCClient) FuzzSearchUserByName(name string) ([]User, error) {
	if name == "" {
		return nil, nil
	}
	if c.oryEnabled() {
		return getUserByKey(c.oryKratosPrivateAddr(), name)
	}
	query := fmt.Sprintf("username:%s OR nickname:%s", name, name)

	return c.findUsersByQuery(query)
}

func userPagingListMapper(user *userPaging) []User {
	userList := make([]User, 0)
	for _, u := range user.Data {
		userList = append(userList, User{
			ID:        strutil.String(u.Id),
			Name:      u.Username,
			Nick:      u.Nickname,
			AvatarURL: u.Avatar,
			Phone:     u.Mobile,
			Email:     u.Email,
		})
	}
	return userList
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

	var user UcUser
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
	if user.ID == 0 {
		return nil, errors.Errorf("failed to find user %s", userID)
	}

	return userMapper(&user), nil
}

func userMapper(user *UcUser) *User {
	return &User{
		ID:        strconv.Itoa(user.ID),
		Name:      user.Name,
		Nick:      user.Nick,
		AvatarURL: user.AvatarURL,
		Phone:     user.Phone,
		Email:     user.Email,
	}
}

func (c *UCClient) findUsersByQuery(query string, idOrder ...string) ([]User, error) {
	// 获取token
	token, err := c.ucTokenAuth.GetServerToken(false)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get token when finding users")
	}

	// 批量查询用户
	var users []UcUser
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
			userMap[strconv.Itoa(user.ID)] = *userMapper(&user)
		}
		var orderedUsers []User
		for _, id := range idOrder {
			orderedUsers = append(orderedUsers, userMap[id])
		}
		return orderedUsers, nil
	}

	userList := make([]User, len(users))
	for i, user := range users {
		userList[i] = *userMapper(&user)
	}

	return userList, nil
}
