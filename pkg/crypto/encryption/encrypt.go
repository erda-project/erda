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

package encryption

import (
	"strings"
)

// Encrypt encrypt 实例对象封装
type EnvEncrypt struct {
	rsaScrypt *RsaCrypt
}

// Option Migration 实例对象配置选项
type Option func(*EnvEncrypt)

// New 新建 Encrypt service
func New(options ...Option) *EnvEncrypt {
	var encrypt EnvEncrypt
	for _, op := range options {
		op(&encrypt)
	}

	return &encrypt
}

// WithRSAScrypt 配置 RsaCrypt
func WithRSAScrypt(rsaScrypt *RsaCrypt) Option {
	return func(a *EnvEncrypt) {
		a.rsaScrypt = rsaScrypt
	}
}

// encryptConfigMap 解密环境变量中password信息
func (e *EnvEncrypt) DecryptAddonConfigMap(configMap *map[string]interface{}) error {
	if len(*configMap) == 0 {
		return nil
	}
	for k, v := range *configMap {
		if strings.Contains(strings.ToLower(k), "pass") || strings.Contains(strings.ToLower(k), "secret") {
			password := v.(string)
			passValue, err := e.rsaScrypt.Decrypt(password, Base64)
			if err != nil {
				return err
			}
			(*configMap)[k] = passValue
		}
	}
	return nil
}

// encryptConfigMap 解密
func (e *EnvEncrypt) DecryptPassword(src string) (string, error) {
	return e.rsaScrypt.Decrypt(src, Base64)
}

// encryptConfigMap 加密
func (e *EnvEncrypt) EncryptPassword(src string) (string, error) {
	return e.rsaScrypt.Encrypt(src, Base64)
}
