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
