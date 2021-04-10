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

package exdto

import "github.com/pkg/errors"

// AuthType
const (
	AT_KEY_AUTH   = "key-auth"
	AT_OAUTH2     = "oauth2"
	AT_SIGN_AUTH  = "sign-auth"
	AT_ALIYUN_APP = "aliyun-app"
)

// AclType
const (
	ACL      = "acl"
	ACL_NONE = ""
	ACL_ON   = "on"
	ACL_OFF  = "off"
)

// Scene
const (
	OPENAPI_SCENE = "openapi"
	WEBAPI_SCENE  = "webapi"
	UNITY_SCENE   = "unity"
)

type PackageDto struct {
	Name             string   `json:"name"`
	BindDomain       []string `json:"bindDomain"`
	AuthType         string   `json:"authType"`
	AclType          string   `json:"aclType"`
	Scene            string   `json:"scene"`
	Description      string   `json:"description"`
	NeedBindCloudapi bool     `json:"needBindCloudapi"`
}

func (dto PackageDto) CheckValid() error {
	if dto.Name == "" || len(dto.BindDomain) == 0 || dto.Scene == "" {
		goto failed
	}
	if !dto.NeedBindCloudapi && dto.AuthType == AT_ALIYUN_APP {
		goto failed
	}
	return nil
failed:
	return errors.Errorf("invalid dto: %+v", dto)
}
