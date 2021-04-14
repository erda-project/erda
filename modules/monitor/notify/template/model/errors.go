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

package model

import "github.com/erda-project/erda-infra/providers/legacy/httpendpoints/errorresp"

var (
	ErrCreateNotify = err("ErrCreateNotify", "创建通知失败")
	ErrGetNotify    = err("ErrGetNotify", "获取通知失败")
	ErrDeleteNotify = err("ErrDeleteNotify", "删除通知失败")
	ErrUpdateNotify = err("ErrUpdateNotify", "更新通知失败")
	ErrNotifyEnable = err("ErrNotifyEnable", "启用通知失败")
)

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}
