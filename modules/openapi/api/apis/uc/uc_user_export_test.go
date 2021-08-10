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

package uc

import (
	"encoding/json"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/ucauth"
)

func TestGetLoginMethodMap(t *testing.T) {
	monkey.Patch(handleListLoginMethod, func(ucauth.OAuthToken) (*listLoginTypeResult, error) {
		return &listLoginTypeResult{RegistryType: []string{"mobile"}}, nil
	})
	defer monkey.UnpatchAll()

	logM, err := getLoginMethodMap(ucauth.OAuthToken{}, "zh-CN")
	assert.NoError(t, err)
	assert.Equal(t, "默认登录方式", logM[""])
}

var fakeUserData = []apistructs.UserInfoExt{{
	UserInfo: apistructs.UserInfo{
		ID:          "1000001",
		Name:        "Hanmeimei",
		Nick:        "meimeiHan",
		Avatar:      "",
		Phone:       "",
		Email:       "fake@fake.com",
		LastLoginAt: "2006-01-02 15:04:05",
		PwdExpireAt: "2006-01-02 15:04:05",
		Source:      "",
	},
}, {
	UserInfo: apistructs.UserInfo{
		ID:          "1000002",
		Name:        "Lilei",
		Nick:        "leiLi",
		Avatar:      "",
		Phone:       "11111111111",
		Email:       "fake1@fake1.com",
		LastLoginAt: "2006-01-02 15:04:05",
		PwdExpireAt: "2006-01-02 15:04:05",
		Source:      "",
	},
},
}

func TestConvertUserToExcelList(t *testing.T) {
	defer monkey.UnpatchAll()
	monkey.Patch(handleListLoginMethod, func(ucauth.OAuthToken) (*listLoginTypeResult, error) {
		return &listLoginTypeResult{RegistryType: []string{"mobile"}}, nil
	})

	logM, err := getLoginMethodMap(ucauth.OAuthToken{}, "zh-CN")

	result := convertUserToExcelList(fakeUserData, logM)

	var expectUserToExcelList = [][]string{
		{"用户名", "昵称", "邮箱", "手机号", "登录方式", "上次登录时间", "密码过期时间", "状态"},
		{"Hanmeimei", "meimeiHan", "fake@fake.com", "", "默认登录方式", "2006-01-02 15:04:05", "2006-01-02 15:04:05", "未冻结"},
		{"Lilei", "leiLi", "fake1@fake1.com", "11111111111", "默认登录方式", "2006-01-02 15:04:05", "2006-01-02 15:04:05", "未冻结"},
	}
	assert.NoError(t, err)
	for i, v := range result {
		for i1, v1 := range v {
			assert.Equal(t, v1, expectUserToExcelList[i][i1])
		}
	}
}

func TestConvertUserToExcelListWithRoles(t *testing.T) {
	var bdl *bundle.Bundle
	var orgMemberFakeData string = `{"organization": {"testorg": {"testprj": ["testapp"]}},"memberList": {"1000001": {"app": {"testapp": {"userId": "1000001","email": "fake@fake.com","mobile": "","name": "Hanmeimei","nick": "meimeiHan","avatar": "","status": "","scope": {"type": "app","id": "10"},"roles": ["Lead"],"labels": null,"removed": false,"deleted": false}},"org": {"testorg": {"userId": "1000001","email": "fake@fake.com","mobile": "","name": "Hanmeimei","nick": "meimeiHan","avatar": "","status": "","scope": {"type": "org","id": "1"},"roles": ["Dev"],"labels": ["Outsource"],"removed": false,"deleted": false}},"prj": {"testprj": {"userId": "1000001","email": "fake@fake.com","mobile": "","name": "Hanmeimei","nick": "meimeiHan","avatar": "","status": "","scope": {"type": "project","id": "5"},"roles": ["Lead"],"labels": null,"removed": false,"deleted": false}}},"1000002": {"app": {"testapp": {"userId": "1000002","email": "fake1@fake1.com","mobile": "","name": "LiLei","nick": "LeiLi","avatar": "","status": "","scope": {"type": "app","id": "10"},"roles": ["Lead"],"labels": null,"removed": false,"deleted": false}},"org": {"testorg": {"userId": "1000002","email": "fake1@fake1.com","mobile": "","name": "LiLei","nick": "LeiLi","avatar": "","status": "","scope": {"type": "org","id": "1"},"roles": ["Dev"],"labels": ["Outsource"],"removed": false,"deleted": false}},"prj": {"testprj": {"userId": "1000002","email": "fake1@fake1.com","mobile": "","name": "LiLei","nick": "LeiLi","avatar": "","status": "","scope": {"type": "project","id": "5"},"roles": ["Lead"],"labels": null,"removed": false,"deleted": false}}}}}`

	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetAllOrganizational",
		func(_ *bundle.Bundle) (*apistructs.GetAllOrganizationalData, error) {
			var organizationalResp apistructs.GetAllOrganizationalData
			if err := json.Unmarshal([]byte(orgMemberFakeData), &organizationalResp); err != nil {
				return nil, err
			}
			return &organizationalResp, nil
		},
	)
	monkey.Patch(handleListLoginMethod, func(ucauth.OAuthToken) (*listLoginTypeResult, error) {
		return &listLoginTypeResult{RegistryType: []string{"mobile"}}, nil
	})

	logM, err := getLoginMethodMap(ucauth.OAuthToken{}, "zh-CN")

	result, err := convertUserToExcelListWithRoles(fakeUserData, logM)
	var expectUserToExcelList = [][]string{
		{"用户名", "昵称", "邮箱", "手机号", "登录方式", "上次登录时间", "密码过期时间", "状态", "企业", "企业角色", "项目", "项目角色", "应用", "应用角色"},
		{"Hanmeimei", "meimeiHan", "fake@fake.com", "", "默认登录方式", "2006-01-02 15:04:05", "2006-01-02 15:04:05", "未冻结", "testorg", "Dev", "testprj", "Lead", "testapp", "Lead"},
		{"Lilei", "leiLi", "fake1@fake1.com", "11111111111", "默认登录方式", "2006-01-02 15:04:05", "2006-01-02 15:04:05", "未冻结", "testorg", "Dev", "testprj", "Lead", "testapp", "Lead"},
	}
	assert.NoError(t, err)
	for i, v := range result {
		for i1, v1 := range v {
			assert.Equal(t, v1, expectUserToExcelList[i][i1])
		}
	}
}
