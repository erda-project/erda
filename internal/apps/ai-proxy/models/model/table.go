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

package model

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/model_type"
)

type Model struct {
	common.BaseModel
	Name       string               `gorm:"column:name;type:varchar(191)" json:"name" yaml:"name"`
	Desc       string               `gorm:"column:desc;type:varchar(1024)" json:"desc" yaml:"desc"`
	Type       model_type.ModelType `gorm:"column:type;type:varchar(32)" json:"type" yaml:"type"`
	Publisher  string               `gorm:"column:publisher;type:varchar(36)" json:"publisher" yaml:"publisher"`
	ProviderID string               `gorm:"column:provider_id;type:char(36)" json:"providerID" yaml:"providerID"`
	ClientID   string               `gorm:"column:client_id;type:char(36)" json:"clientID" yaml:"clientID"`
	APIKey     string               `gorm:"column:api_key;type:varchar(191)" json:"aPIKey" yaml:"aPIKey"`
	Metadata   metadata.Metadata    `gorm:"column:metadata;type:mediumtext" json:"metadata" yaml:"metadata"`

	TemplateID     string            `gorm:"column:template_id;type:varchar(36)" json:"templateID" yaml:"templateID"`
	TemplateParams map[string]string `gorm:"column:template_params;type:jsonb;serializer:json" json:"templateParams" yaml:"templateParams"`
	IsEnabled      *bool             `gorm:"column:is_enabled;type:bool;not null;default:1" json:"isEnabled" yaml:"isEnabled"`
}

func (*Model) TableName() string { return "ai_proxy_model" }

func (m *Model) ToProtobuf() *pb.Model {
	pbMetadata := m.Metadata.ToProtobuf()
	return &pb.Model{
		Id:             m.ID.String,
		CreatedAt:      timestamppb.New(m.CreatedAt),
		UpdatedAt:      timestamppb.New(m.UpdatedAt),
		DeletedAt:      timestamppb.New(m.DeletedAt.Time),
		Name:           m.Name,
		Desc:           m.Desc,
		Type:           pb.ModelType(pb.ModelType_value[string(m.Type)]),
		ProviderId:     m.ProviderID,
		ApiKey:         m.APIKey,
		ClientId:       m.ClientID,
		Publisher:      m.Publisher,
		TemplateId:     m.TemplateID,
		TemplateParams: m.TemplateParams,
		IsEnabled:      m.IsEnabled,
		Metadata:       pbMetadata,
	}
}

type Models []*Model

func (models Models) ToProtobuf() []*pb.Model {
	var pbModels []*pb.Model
	for _, c := range models {
		pbModels = append(pbModels, c.ToProtobuf())
	}
	return pbModels
}
