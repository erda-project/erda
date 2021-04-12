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

package apierrors

import "github.com/erda-project/erda/pkg/httpserver/errorresp"

var (
	// uc users
	ErrListUser                = err("ErrListUser", "获取用户列表失败")
	ErrGetUser                 = err("ErrGetUser", "获取用户详情失败")
	ErrCreateUser              = err("ErrCreateUser", "创建用户失败")
	ErrUpdateUserInfo          = err("ErrUpdateUserInfo", "更新用户信息失败")
	ErrFreezeUser              = err("ErrFreezeUser", "冻结用户失败")
	ErrBatchFreezeUser         = err("ErrBatchFreezeUser", "批量冻结用户失败")
	ErrUnfreezeUser            = err("ErrUnfreezeUser", "解冻用户失败")
	ErrBatchUnFreezeUser       = err("ErrBatchUnFreezeUser", "批量解冻用户失败")
	ErrGetPwdSecurityConfig    = err("ErrGetPwdSecurityConfig", "获取密码安全配置失败")
	ErrUpdatePwdSecurityConfig = err("ErrUpdatePwdSecurityConfig", "更新密码安全配置失败")
	ErrUpdateLoginMethod       = err("ErrUpdateLoginMethod", "更新用户登录方式失败")
	ErrListLoginMethod         = err("ErrListLoginMethod", "获取用户登录方式失败")
	ErrAdminUser               = err("ErrAdminUser", "用户操作失败")
)

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}
