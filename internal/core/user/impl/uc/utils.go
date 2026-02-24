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
	"strconv"

	"github.com/samber/lo"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/user/pb"
)

func userMapper(user *GetUser) *commonpb.UserInfo {
	return &commonpb.UserInfo{
		Id:     strconv.Itoa(user.ID),
		Name:   user.Name,
		Nick:   user.Nick,
		Avatar: user.AvatarURL,
		Phone:  user.Phone,
		Email:  user.Email,
	}
}

func usersMapper(users []*GetUser) []*commonpb.UserInfo {
	return lo.Map(users, func(item *GetUser, _ int) *commonpb.UserInfo {
		return userMapper(item)
	})
}

func managedUserMapper(u *UserDto) *pb.ManagedUser {
	var lastLoginAt *timestamppb.Timestamp
	if u.LastLoginAt != nil && !u.LastLoginAt.IsZero() {
		lastLoginAt = timestamppb.New(u.LastLoginAt.Time)
	}

	var pwdExpireAt *timestamppb.Timestamp
	if u.PwdExpireAt != nil && !u.PwdExpireAt.IsZero() {
		pwdExpireAt = timestamppb.New(u.PwdExpireAt.Time)
	}

	return &pb.ManagedUser{
		Id:          cast.ToString(u.Id),
		Name:        u.Username,
		Nick:        u.Nickname,
		Avatar:      u.Avatar,
		Phone:       u.Mobile,
		Email:       u.Email,
		LastLoginAt: lastLoginAt,
		PwdExpireAt: pwdExpireAt,
		Source:      u.Source,
		Locked:      u.Locked,
	}
}
