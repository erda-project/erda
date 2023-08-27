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
)

type Model struct {
	common.BaseModel
	Name       string            `gorm:"column:name;type:varchar(191)" json:"name" yaml:"name"`
	Desc       string            `gorm:"column:desc;type:varchar(1024)" json:"desc" yaml:"desc"`
	Type       ModelType         `gorm:"column:type;type:varchar(32)" json:"type" yaml:"type"`
	ProviderID string            `gorm:"column:provider_id;type:char(36)" json:"providerID" yaml:"providerID"`
	APIKey     string            `gorm:"column:api_key;type:varchar(128)" json:"aPIKey" yaml:"aPIKey"`
	Metadata   metadata.Metadata `gorm:"column:metadata;type:mediumtext" json:"metadata" yaml:"metadata"`
}

func (*Model) TableName() string { return "ai_proxy_model" }

// see: api/proto/apps/aiproxy/model/model.proto#ModelType
type ModelType string

func GetModelTypeFromProtobuf(pbModelType pb.ModelType) ModelType {
	return ModelType(pb.ModelType_name[int32(pbModelType)])
}

func (m *Model) ToProtobuf() *pb.Model {
	return &pb.Model{
		Id:         m.ID.String,
		CreatedAt:  timestamppb.New(m.CreatedAt),
		UpdatedAt:  timestamppb.New(m.UpdatedAt),
		DeletedAt:  timestamppb.New(m.DeletedAt.Time),
		Name:       m.Name,
		Desc:       m.Desc,
		Type:       pb.ModelType(pb.ModelType_value[string(m.Type)]),
		ProviderId: m.ProviderID,
		ApiKey:     m.APIKey,
		Metadata:   m.Metadata.ToProtobuf(),
	}
}
