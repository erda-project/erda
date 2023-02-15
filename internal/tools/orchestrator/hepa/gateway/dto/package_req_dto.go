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

package dto

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
)

// AuthType
const (
	AT_KEY_AUTH   = "key-auth"
	AT_OAUTH2     = "oauth2"
	AT_SIGN_AUTH  = "sign-auth"
	AT_HMAC_AUTH  = "hmac-auth"
	AT_ALIYUN_APP = "aliyun-app"
)

// AclType
const (
	ACL      = "acl"
	ACL_NONE = ""
	ACL_ON   = "on"
	ACL_OFF  = "off"
)

type PackageDto struct {
	Name             string   `json:"name"`
	BindDomain       []string `json:"bindDomain"`
	AuthType         string   `json:"authType"`
	AclType          string   `json:"aclType"`
	Scene            string   `json:"scene"`
	Description      string   `json:"description"`
	NeedBindCloudapi bool     `json:"needBindCloudapi"`
	GatewayProvider  string   `json:"gatewayProvider"` // "MSE" 或 ""
}

func (dto PackageDto) CheckValid() error {
	switch dto.Scene {
	case orm.UnityScene, orm.HubScene, orm.WebapiScene, orm.OpenapiScene:
	default:
		return errors.Errorf("invalid scene: %s", dto.Scene)
	}
	switch 0 {
	case len(dto.Name):
		return errors.Errorf("invalid name: %s", "empty")
	case len(dto.BindDomain):
		return errors.Errorf("invalid bind domain: %s", "no bind domain")
	}
	if !dto.NeedBindCloudapi && dto.AuthType == AT_ALIYUN_APP {
		return errors.Errorf("invalid parameter: %s", AT_ALIYUN_APP+" need bind cloud api")
	}
	return nil
}
