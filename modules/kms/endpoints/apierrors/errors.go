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

// Package apierrors defines all service errors.
package apierrors

import (
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

var (
	ErrCheckIdentity    = err("ErrCheckIdentity", "身份校验失败")
	ErrParseRequest     = err("ErrParseRequest", "解析请求失败")
	ErrCreateKey        = err("ErrCreateKey", "创建 KMS 用户主密钥失败")
	ErrEncrypt          = err("ErrEncrypt", "对称加密失败")
	ErrDecrypt          = err("ErrDecrypt", "对称解密失败")
	ErrGenerateDataKey  = err("ErrGenerateDataKey", "生成数据加密密钥失败")
	ErrRotateKeyVersion = err("ErrRotateKeyVersion", "轮转密钥版本失败")
	ErrDescribeKey      = err("ErrDescribeKey", "查询用户主密钥失败")
)

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}
