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

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}

var (
	ErrCreateExtension        = err("ErrCreateExtension", "添加扩展失败")
	ErrQueryExtension         = err("ErrQueryExtension", "查询扩展失败")
	ErrCreateExtensionVersion = err("ErrCreateExtensionVersion", "添加扩展版本失败")
	ErrQueryExtensionVersion  = err("ErrQueryExtensionVersion", "查询扩展版本失败")
	ErrDeleteExtensionVersion = err("ErrDeleteExtensionVersion", "删除扩展版本失败")
)
