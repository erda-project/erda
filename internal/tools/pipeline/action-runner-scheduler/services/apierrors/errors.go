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

package apierrors

import (
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
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
