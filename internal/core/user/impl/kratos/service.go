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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/pkg/desensitize"
	"github.com/erda-project/erda/pkg/strutil"
)

func (p *provider) oryKratosPrivateAddr() string {
	return p.baseURL
}

func (p *provider) FindUsers(ids []string) ([]common.User, error) {
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
	return users, nil
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
func (p *provider) FindUsersByKey(key string) ([]common.User, error) {
	if key == "" {
		return nil, nil
	}
	return getUserByKey(p.oryKratosPrivateAddr(), key)
}

// GetUser 获取用户详情
func (p *provider) GetUser(userID string) (*common.User, error) {
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
	return user, nil
}

func (p *provider) GetUsers(IDs []string, needDesensitize bool) (map[string]apistructs.UserInfo, error) {
	b, err := p.FindUsers(IDs)
	if err != nil {
		return nil, err
	}

	users := make(map[string]apistructs.UserInfo, len(b))
	if needDesensitize {
		for i := range b {
			users[string(b[i].ID)] = apistructs.UserInfo{
				ID:     "",
				Email:  desensitize.Email(b[i].Email),
				Phone:  desensitize.Mobile(b[i].Phone),
				Avatar: b[i].AvatarURL,
				Name:   desensitize.Name(b[i].Name),
				Nick:   desensitize.Name(b[i].Nick),
			}
		}
	} else {
		for i := range b {
			users[string(b[i].ID)] = apistructs.UserInfo{
				ID:     string(b[i].ID),
				Email:  b[i].Email,
				Phone:  b[i].Phone,
				Avatar: b[i].AvatarURL,
				Name:   b[i].Name,
				Nick:   b[i].Nick,
			}
		}
	}
	for _, userID := range IDs {
		_, exist := users[userID]
		if !exist {
			users[userID] = apistructs.UserInfo{
				ID:     userID,
				Email:  "",
				Phone:  "",
				Avatar: "",
				Name:   "用户已注销",
				Nick:   "用户已注销",
			}
		}
	}
	return users, nil
}
