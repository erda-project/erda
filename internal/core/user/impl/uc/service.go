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

//
//import (
//	"bytes"
//	"context"
//	"encoding/json"
//	"fmt"
//	"net/http"
//	"strconv"
//	"strings"
//
//	"github.com/pkg/errors"
//
//	"github.com/erda-project/erda-proto-go/core/user/pb"
//	"github.com/erda-project/erda/internal/core/user/common"
//	"github.com/erda-project/erda/pkg/strutil"
//)
//
//// User 用户中心用户数据结构
//type UcUser struct {
//	ID        int    `json:"user_id"`
//	Name      string `json:"username"`
//	Nick      string `json:"nickname"`
//	AvatarURL string `json:"avatar_url"`
//	Phone     string `json:"phone_number"`
//	Email     string `json:"email"`
//}
//
//func (p *provider) FindUsers(ctx context.Context, req *pb.FindUsersRequest) (*pb.FindUsersResponse, error) {
//	ids := req.IDs
//	if len(ids) == 0 {
//		return &pb.FindUsersResponse{}, nil
//	}
//	sysOpExist := strutil.Exist(ids, common.SystemOperator)
//	if sysOpExist {
//		ids = strutil.RemoveSlice(ids, common.SystemOperator)
//	}
//
//	parts := make([]string, 0)
//	for _, id := range ids {
//		parts = append(parts, "user_id:"+id)
//	}
//	query := strings.Join(parts, " OR ")
//	users, err := p.findUsersByQuery(query, ids...)
//	if err != nil {
//		return nil, err
//	}
//	if sysOpExist {
//		users = append(users, common.SystemUser)
//	}
//	userList := make([]*pb.User, 0, len(users))
//	for _, i := range users {
//		userList = append(userList, common.ToPbUser(i))
//	}
//	return &pb.FindUsersResponse{Data: userList}, nil
//}
//
//// FindUsersByKey 根据key查找用户，key可匹配用户名/邮箱/手机号
//func (p *provider) FindUsersByKey(ctx context.Context, req *pb.FindUsersByKeyRequest) (*pb.FindUsersByKeyResponse, error) {
//	key := req.Key
//	if key == "" {
//		return &pb.FindUsersByKeyResponse{}, nil
//	}
//	query := fmt.Sprintf("username:%s OR nickname:%s OR phone_number:%s OR email:%s", key, key, key, key)
//	users, err := p.findUsersByQuery(query)
//	if err != nil {
//		return nil, err
//	}
//	userList := make([]*pb.User, 0, len(users))
//	for _, i := range users {
//		userList = append(userList, common.ToPbUser(i))
//	}
//	return &pb.FindUsersByKeyResponse{Data: userList}, nil
//}
//
//// GetUser 获取用户详情
//func (p *provider) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
//	var (
//		user *common.User
//		err  error
//	)
//	userID := req.UserID
//	// 增加重试机制，防止因 uc 升级 serverToken 格式不兼容，无法获取用户信息
//	for i := 0; i < 3; i++ {
//		user, err = p.getUser(userID)
//		if err != nil {
//			continue
//		}
//		return &pb.GetUserResponse{
//			Data: common.ToPbUser(*user),
//		}, nil
//	}
//
//	return nil, err
//}
//
//func (p *provider) findUsersByQuery(query string, idOrder ...string) ([]common.User, error) {
//	token, err := p.UserOAuthSvc.ExchangeClientCredentials(context.Background(), false, nil)
//	if err != nil {
//		return nil, errors.Wrapf(err, "failed to get token when finding users")
//	}
//
//	// 批量查询用户
//	var users []UcUser
//	var b bytes.Buffer
//	r, err := p.client.Get(p.baseURL).
//		Path("/api/open/v1/users").
//		Param("query", query).
//		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).
//		Do().Body(&b)
//	if err != nil {
//		return nil, errors.Wrapf(err, "failed to find users")
//	}
//	if !r.IsOK() {
//		return nil, errors.Errorf("failed to find users, status code: %d", r.StatusCode())
//	}
//	content := b.Bytes()
//	if err := json.Unmarshal(content, &users); err != nil {
//		return nil, fmt.Errorf("failed to unmarshal: %v", string(content))
//	}
//
//	// 保证顺序
//	if len(idOrder) > 0 {
//		userMap := make(map[string]common.User)
//		for _, user := range users {
//			userMap[strconv.Itoa(user.ID)] = *userMapper(&user)
//		}
//		var orderedUsers []common.User
//		for _, id := range idOrder {
//			orderedUsers = append(orderedUsers, userMap[id])
//		}
//		return orderedUsers, nil
//	}
//
//	userList := make([]common.User, len(users))
//	for i, user := range users {
//		userList[i] = *userMapper(&user)
//	}
//
//	return userList, nil
//}
//
//func userMapper(user *UcUser) *common.User {
//	return &common.User{
//		ID:        strconv.Itoa(user.ID),
//		Name:      user.Name,
//		Nick:      user.Nick,
//		AvatarURL: user.AvatarURL,
//		Phone:     user.Phone,
//		Email:     user.Email,
//	}
//}
//
//func (p *provider) getUser(userID string) (*common.User, error) {
//	token, err := p.UserOAuthSvc.ExchangeClientCredentials(context.Background(), false, nil)
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to get token when get user")
//	}
//
//	var user UcUser
//	r, err := p.client.Get(p.baseURL).
//		Path(strutil.Concat("/api/open/v1/users/", userID)).
//		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).
//		Do().JSON(&user)
//	if err != nil {
//		return nil, err
//	}
//
//	if !r.IsOK() {
//		if r.StatusCode() == http.StatusUnauthorized {
//			//p.InvalidateServerToken()
//		}
//		return nil, errors.Errorf("failed to find user, status code: %d", r.StatusCode())
//	}
//	if user.ID == 0 {
//		return nil, errors.Errorf("failed to find user %s", userID)
//	}
//
//	return userMapper(&user), nil
//}
