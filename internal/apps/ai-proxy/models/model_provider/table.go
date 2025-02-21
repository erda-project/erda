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

package model_provider

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/model_provider_type"
)

type ModelProvider struct {
	common.BaseModel
	Name     string                                `gorm:"column:name;type:varchar(191)" json:"name" yaml:"name"`
	Desc     string                                `gorm:"column:desc;type:varchar(1024)" json:"desc" yaml:"desc"`
	Type     model_provider_type.ModelProviderType `gorm:"column:type;type:varchar(191)" json:"type" yaml:"type"`
	APIKey   string                                `gorm:"column:api_key;type:varchar(191)" json:"aPIKey" yaml:"aPIKey"`
	Metadata metadata.Metadata                     `gorm:"column:metadata;type:mediumtext" json:"metadata" yaml:"metadata"`
}

func (*ModelProvider) TableName() string { return "ai_proxy_model_provider" }

func (m *ModelProvider) ToProtobuf() *pb.ModelProvider {
	return &pb.ModelProvider{
		Id:        m.ID.String,
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
		DeletedAt: timestamppb.New(m.DeletedAt.Time),
		Name:      m.Name,
		Desc:      m.Desc,
		Type:      string(m.Type),
		ApiKey:    m.APIKey,
		Metadata:  m.Metadata.ToProtobuf(),
	}
}

type ModelProviders []*ModelProvider

func (modelProviders ModelProviders) ToProtobuf() []*pb.ModelProvider {
	var pbClients []*pb.ModelProvider
	for _, c := range modelProviders {
		pbClients = append(pbClients, c.ToProtobuf())
	}
	return pbClients
}
