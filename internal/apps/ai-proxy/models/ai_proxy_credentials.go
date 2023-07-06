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

	AccessKeyId string    `json:"accessKeyId" yaml:"accessKeyId" gorm:"access_key_id"`
	SecretKeyId string    `json:"secretKeyId" yaml:"secretKeyId" gorm:"secret_key_id"`
	Name        string    `json:"name" yaml:"name" gorm:"name"`
	Platform    string    `json:"platform" yaml:"platform" gorm:"platform"`
	Description string    `json:"description" yaml:"description" gorm:"description"`
	Enabled     bool      `json:"enabled" yaml:"enabled" gorm:"enabled"`
	ExpiredAt   time.Time `json:"expiredAt" yaml:"expiredAt" gorm:"expired_at"`
}

func (AIProxyCredentials) TableName() string {
	return "ai_proxy_credentials"
}

func (model AIProxyCredentials) ToProtobuf() *pb.Credential {
	return &pb.Credential{
		AccessKeyId: model.AccessKeyId,
		SecretKeyId: model.SecretKeyId,
		Name:        model.Name,
		Platform:    model.Platform,
		Description: model.Description,
		Enabled:     model.Enabled,
	}
}
