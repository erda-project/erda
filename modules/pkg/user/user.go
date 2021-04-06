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

// Package user 定义通用的 user 逻辑.
package user

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httputil"
)

// ID 定义 user id.
type ID string

// String 返回 user id 字符串类型.
func (i ID) String() string {
	return string(i)
}

// Invalid 返回 user id 是否有效.
func (i ID) Invalid() bool {
	return string(i) == ""
}

// GetUserID 从 http request 的 header 中读取 user id.
func GetUserID(r *http.Request) (ID, error) {
	v := r.Header.Get("USER-ID")
	id := ID(v)

	if id.Invalid() {
		return id, errors.New("invalid user id")
	}
	return id, nil
}

// GetIdentityInfo 从 http.Request 中获取用户 ID && Internal-Client
//
// return: IdentityInfo, error
func GetIdentityInfo(r *http.Request) (apistructs.IdentityInfo, error) {
	// 尝试从 Header 中获取用户信息
	headerUserID, headerUserErr := GetUserID(r)

	// 尝试从 Header 中获取 Internal-Client
	internalClient := r.Header.Get(httputil.InternalHeader)

	// 未登录
	if headerUserErr != nil && internalClient == "" {
		return apistructs.IdentityInfo{}, errors.Errorf("invalid identity info")
	}

	identity := apistructs.IdentityInfo{
		UserID:         string(headerUserID),
		InternalClient: internalClient,
	}

	return identity, nil
}
