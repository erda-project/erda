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

package service_provider

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

type ServiceProvider struct {
	common.BaseModel
	Name     string            `gorm:"column:name;type:varchar(191)" json:"name" yaml:"name"`
	Desc     string            `gorm:"column:desc;type:varchar(1024)" json:"desc" yaml:"desc"`
	Type     string            `gorm:"column:type;type:varchar(191)" json:"type" yaml:"type"`
	APIKey   string            `gorm:"column:api_key;type:varchar(191)" json:"apiKey" yaml:"apiKey"`
	ClientID string            `gorm:"column:client_id;type:char(36)" json:"clientId" yaml:"clientId"`
	Metadata metadata.Metadata `gorm:"column:metadata;type:mediumtext" json:"metadata" yaml:"metadata"`
}

func (*ServiceProvider) TableName() string { return "ai_proxy_model_provider" }

func (m *ServiceProvider) ToProtobuf() *pb.ServiceProvider {
	return &pb.ServiceProvider{
		Id:        m.ID.String,
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
		DeletedAt: timestamppb.New(m.DeletedAt.Time),
		Name:      m.Name,
		Desc:      m.Desc,
		Type:      m.Type,
		ApiKey:    m.APIKey,
		ClientId:  m.ClientID,
		Metadata:  m.Metadata.ToProtobuf(),
	}
}

type ServiceProviders []*ServiceProvider

func (serviceProviders ServiceProviders) ToProtobuf() []*pb.ServiceProvider {
	var pbClients []*pb.ServiceProvider
	for _, c := range serviceProviders {
		pbClients = append(pbClients, c.ToProtobuf())
	}
	return pbClients
}
