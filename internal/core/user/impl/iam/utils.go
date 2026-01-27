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

package iam

import (
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/user/pb"
)

const timeLayout = "2006-01-02T15:04:05"

func formatIAMTime(timeStr string) (time.Time, error) {
	return time.Parse(timeLayout, timeStr)
}

func userMapper(user *UserDto) *commonpb.UserInfo {
	return &commonpb.UserInfo{
		Id:     cast.ToString(user.ID),
		Name:   user.Username,
		Nick:   user.Nickname,
		Avatar: user.Avatar,
		Phone:  user.Mobile,
		Email:  user.Email,
	}
}

func managedUserMapper(u *UserDto) (*pb.ManagedUser, error) {
	var (
		loginAt          *timestamppb.Timestamp
		passwordExpireAt *timestamppb.Timestamp
	)

	if u.LastLoginAt != "" {
		parsedLoginAt, err := formatIAMTime(u.LastLoginAt)
		if err != nil {
			return nil, err
		}
		loginAt = timestamppb.New(parsedLoginAt)
	}

	if u.PasswordExpireAt != "" {
		parsedPasswordExpireAt, err := formatIAMTime(u.PasswordExpireAt)
		if err != nil {
			return nil, err
		}
		passwordExpireAt = timestamppb.New(parsedPasswordExpireAt)
	}

	return &pb.ManagedUser{
		Id:          strconv.FormatInt(u.ID, 10),
		Name:        u.Username,
		Nick:        u.Nickname,
		Avatar:      u.Avatar,
		Phone:       u.Mobile,
		Email:       u.Email,
		LastLoginAt: loginAt,
		PwdExpireAt: passwordExpireAt,
		// Not support source change now.
		//Source: u.Source,
		Locked: u.Locked,
	}, nil
}

func isEmptyTrim(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
