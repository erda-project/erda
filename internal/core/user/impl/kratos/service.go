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

package kratos

import (
	"context"

	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/pkg/strutil"
)

func (p *provider) oryKratosPrivateAddr() string {
	return p.baseURL
}

func (p *provider) FindUsers(ctx context.Context, req *pb.FindUsersRequest) (*pb.FindUsersResponse, error) {
	ids := req.IDs
	if len(ids) == 0 {
		return nil, nil
	}
	sysOpExist := strutil.Exist(ids, common.SystemOperator)
	if sysOpExist {
		ids = strutil.RemoveSlice(ids, common.SystemOperator)
	}

	// get ordered uuid list
	userIDs, userMap, err := p.ConvertUserIDs(ids)
	if err != nil {
		return nil, err
	}
	users, err := getUserByIDs(p.oryKratosPrivateAddr(), userIDs)
	if err != nil {
		return nil, err
	}
	// revert uuid to id for old uc users
	for i, u := range users {
		if userID, ok := userMap[u.ID]; ok {
			users[i].ID = userID
		}
	}
	if sysOpExist {
		users = append(users, common.SystemUser)
	}
	userList := make([]*pb.User, 0, len(users))
	for _, i := range users {
		userList = append(userList, common.ToPbUser(i))
	}
	return &pb.FindUsersResponse{Data: userList}, nil
}

func (p *provider) ConvertUserIDs(ids []string) ([]string, map[string]string, error) {
	users, err := p.GetUserIDMapping(ids)
	if err != nil {
		return nil, nil, err
	}
	ucKratosMap := make(map[string]string)
	kratosUcMap := make(map[string]string)
	for _, u := range users {
		ucKratosMap[u.ID] = u.UserID
		kratosUcMap[u.UserID] = u.ID
	}
	return filterUserIDs(ids, ucKratosMap), kratosUcMap, nil
}

func filterUserIDs(ids []string, users map[string]string) []string {
	userIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		if userID, ok := users[id]; ok {
			userIDs = append(userIDs, userID)
		} else {
			userIDs = append(userIDs, id)
		}
	}
	return userIDs
}

type UserIDModel struct {
	ID     string
	UserID string
}

func (p *provider) GetUserIDMapping(ids []string) ([]UserIDModel, error) {
	var users []UserIDModel
	if err := p.DB.Table("kratos_uc_userid_mapping").Where("id in (?)", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FindUsersByKey 根据key查找用户，key可匹配用户名/邮箱/手机号
func (p *provider) FindUsersByKey(ctx context.Context, req *pb.FindUsersByKeyRequest) (*pb.FindUsersByKeyResponse, error) {
	key := req.Key
	if key == "" {
		return nil, nil
	}
	users, err := getUserByKey(p.oryKratosPrivateAddr(), key)
	if err != nil {
		return nil, err
	}
	userList := make([]*pb.User, 0, len(users))
	for _, i := range users {
		userList = append(userList, common.ToPbUser(i))
	}
	return &pb.FindUsersByKeyResponse{Data: userList}, nil
}

// GetUser 获取用户详情
func (p *provider) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	userID := req.UserID
	userIDs, userMap, err := p.ConvertUserIDs([]string{userID})
	if err != nil || len(userIDs) == 0 {
		return nil, err
	}
	user, err := getUserByID(p.oryKratosPrivateAddr(), userIDs[0])
	if err != nil {
		return nil, err
	}
	if userID, ok := userMap[user.ID]; ok {
		user.ID = userID
	}
	return &pb.GetUserResponse{
		Data: common.ToPbUser(*user),
	}, nil
}
