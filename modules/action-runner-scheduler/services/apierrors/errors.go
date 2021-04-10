//  Copyright (c) 2021 Terminus, Inc.
//
//  This program is free software: you can use, redistribute, and/or modify
//  it under the terms of the GNU Affero General Public License, version 3
//  or later ("AGPL"), as published by the Free Software Foundation.
//
//  This program is distributed in the hope that it will be useful, but WITHOUT
//  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
//  FITNESS FOR A PARTICULAR PURPOSE.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program. If not, see <http://www.gnu.org/licenses/>.

package apierrors

import (
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

var (
	ErrCreateRunnerTask  = err("ErrCreateRunnerTask", "创建runner任务失败")
	ErrGetRunnerTask     = err("ErrGetRunnerTask", "获取runner任务失败")
	ErrUpdateRunnerTask  = err("ErrUpdateRunnerTask", "更新runner任务失败")
	ErrFetchRunnerTask   = err("ErrFetchRunnerTask", "获取runner任务失败")
	ErrCollectRunnerLogs = err("ErrCollectRunnerLogs", "收集runner日志失败")
)

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}
