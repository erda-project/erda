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

package models

import (
	"time"

	"github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
)

type AIProxyCredentials struct {
	Id        fields.UUID      `json:"id" yaml:"id" gorm:"id"`
	CreatedAt time.Time        `json:"createdAt" yaml:"createdAt" gorm:"created_at"`
	UpdatedAt time.Time        `json:"updatedAt" yaml:"updatedAt" gorm:"updated_at"`
	DeletedAt fields.DeletedAt `json:"deletedAt" yaml:"deletedAt" gorm:"deleted_at"`

	AccessKeyID        string    `gorm:"type:char(32);not null;comment:平台接入 AI 服务的 AK"`
	SecretKeyID        string    `gorm:"type:char(32);not null;comment:平台接入 AI 服务的 SK"`
	Name               string    `gorm:"type:varchar(64);not null;comment:凭证名称"`
	Platform           string    `gorm:"type:varchar(128);not null;comment:接入 AI 服务的平台"`
	Description        string    `gorm:"type:varchar(512);not null;comment:凭证描述"`
	Enabled            bool      `gorm:"not null;default:true;comment:是否启用该凭证"`
	ExpiredAt          time.Time `gorm:"not null;default:'2099-01-01 00:00:00';comment:凭证过期时间"`
	ProviderName       string    `gorm:"type:varchar(128);not null;default:'';comment:AI 服务 Provider 名称"`
	ProviderInstanceID string    `gorm:"type:varchar(512);not null;default:'';comment:AI 服务 Provider 实例 id"`
}

func NewCredential(credential *pb.Credential) *AIProxyCredentials {
	var model AIProxyCredentials
	model.AccessKeyID = credential.GetAccessKeyId()
	model.SecretKeyID = credential.GetSecretKeyId()
	model.Name = credential.GetName()
	model.Platform = credential.GetPlatform()
	model.Description = credential.GetDescription()
	model.Enabled = credential.GetEnabled()
	model.ExpiredAt = time.Date(2099, time.January, 1, 0, 0, 0, 0, time.Local)
	model.ProviderName = credential.GetProviderName()
	model.ProviderInstanceID = credential.GetProviderInstanceId()
	return &model
}

func (AIProxyCredentials) TableName() string {
	return "ai_proxy_credentials"
}

func (model AIProxyCredentials) ToProtobuf() *pb.Credential {
	return &pb.Credential{
		AccessKeyId:        model.AccessKeyID,
		SecretKeyId:        model.SecretKeyID,
		Name:               model.Name,
		Platform:           model.Platform,
		Description:        model.Description,
		Enabled:            model.Enabled,
		ProviderName:       model.ProviderName,
		ProviderInstanceId: model.ProviderInstanceID,
	}
}
