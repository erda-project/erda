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

package utils

import (
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

// runtime errors
var (
	ErrCreateRuntimeHPARule        = err("ErrCreateRuntimeHPARule", "创建 Runtime 应用服务 HPA 规则失败")
	ErrDeleteRuntimeHPARule        = err("ErrDeleteRuntimeHPARule", "删除 Runtime 应用服务 HPA 规则失败")
	ErrListRuntimeHPARule          = err("ErrListRuntimeHPARule", "查询 Runtime 应用服务 HPA 规则失败")
	ErrUpdateRuntimeHPARule        = err("ErrUpdateRuntimeHPARule", "更新 Runtime 应用服务 HPA 规则失败")
	ErrApplyOrCancelRuntimeHPARule = err("ErrApplyOrCancelRuntimeHPARule", "开启/关闭 Runtime 应用服务 HPA 功能失败")
	ErrUnknownActionRuntimeHPARule = err("ErrUnknownActionRuntimeHPARule", "暂不支持 Runtime 应用服务 HPA 规则操作类型")
	ErrListRuntimeServiceBaseInfo  = err("ErrListRuntimeServiceBaseInfo", "查询 Runtime 应用服务基准信息失败")
)

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}
