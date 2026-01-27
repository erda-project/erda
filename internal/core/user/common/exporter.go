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

package common

import (
	"bytes"
	"errors"
	"io"

	"github.com/spf13/cast"

	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/excel"
)

func ExportExcel(exporterResp *apis.Response) (io.Reader, string, error) {
	var (
		table     [][]string
		tableName = "users"

		buf = bytes.NewBuffer([]byte{})
	)

	resp, ok := exporterResp.Data.(*pb.UserExportResponse)
	if !ok {
		return nil, "", errors.New("data type error")
	}

	table = convertUserToExcelList(
		convertPbToUserInfoExt(resp.List),
		resp.LoginMethods,
	)

	if err := excel.ExportExcel(buf, table, tableName); err != nil {
		return nil, "", err
	}

	return buf, tableName, nil
}

func convertUserToExcelList(users []*apistructs.UserInfoExt, loginMethodMap map[string]string) [][]string {
	r := [][]string{{"用户名", "昵称", "邮箱", "手机号", "登录方式", "上次登录时间", "密码过期时间", "状态"}}
	for _, user := range users {
		state := "未冻结"
		if user.Locked {
			state = "冻结"
		}
		r = append(r, []string{user.Name, user.Nick, user.Email, user.Phone, loginMethodMap[user.Source], user.LastLoginAt, user.PwdExpireAt, state})
	}
	return r
}

func convertPbToUserInfoExt(users []*pb.ManagedUser) []*apistructs.UserInfoExt {
	userInfoExt := make([]*apistructs.UserInfoExt, 0, len(users))
	for _, u := range users {
		userInfoExt = append(userInfoExt, &apistructs.UserInfoExt{
			UserInfo: apistructs.UserInfo{
				ID:          cast.ToString(u.Id),
				Name:        u.Name,
				Nick:        u.Nick,
				Avatar:      u.Avatar,
				Phone:       u.Phone,
				Email:       u.Email,
				LastLoginAt: u.LastLoginAt.AsTime().Format("2006-01-02 15:04:05"),
				PwdExpireAt: u.PwdExpireAt.AsTime().Format("2006-01-02 15:04:05"),
				Source:      u.Source,
			},
			Locked: u.Locked,
		})
	}
	return userInfoExt
}
