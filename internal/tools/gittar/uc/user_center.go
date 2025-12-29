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

import (
	"context"
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"

	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/desensitize"
)

var userCache = cache.New(12*time.Hour, 1*time.Hour)

var userServiceServer userpb.UserServiceServer

func InitializeUcClient(identity userpb.UserServiceServer) {
	userServiceServer = identity
	logrus.Infof("gittar uc client set up")
}

func FindUserById(id string) (*apistructs.UserInfoDto, error) {
	cacheUser, exist := userCache.Get(id)
	if exist {
		return cacheUser.(*apistructs.UserInfoDto), nil
	}
	if id == "0" {
		return nil, nil
	}
	user, err := userServiceServer.GetUser(context.Background(), &userpb.GetUserRequest{
		UserID: id,
	})
	if err != nil {
		return nil, err
	}
	userInfo := convertDto(user.Data)
	userCache.Set(id, userInfo, cache.DefaultExpiration)
	return userInfo, nil
}
func FindUserByIdWithDesensitize(id string) (*apistructs.UserInfoDto, error) {
	user, err := userServiceServer.GetUser(context.Background(), &userpb.GetUserRequest{
		UserID: id,
	})
	if err != nil {
		return nil, err
	}
	userInfo := user.Data
	if userInfo != nil {
		// 不是用指针类型 防止更新缓存里的原始值
		result := userInfo
		result.Email = desensitize.Email(userInfo.Email)
		result.Phone = desensitize.Mobile(userInfo.Phone)
		userInfo = result
		return convertDto(userInfo), nil
	}
	return nil, errors.New("error user")
}

func convertDto(userInfo *userpb.User) *apistructs.UserInfoDto {
	return &apistructs.UserInfoDto{
		AvatarURL: userInfo.AvatarURL,
		Email:     userInfo.Email,
		UserID:    userInfo.ID,
		NickName:  userInfo.Nick,
		Phone:     userInfo.Phone,
		Username:  userInfo.Name,
	}
}
