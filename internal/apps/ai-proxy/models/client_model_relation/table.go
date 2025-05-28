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

package client_model_relation

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
)

type ClientModelRelation struct {
	common.BaseModel
	ClientID string `gorm:"column:client_id;type:char(36)" json:"clientID" yaml:"clientID"`
	ModelID  string `gorm:"column:model_id;type:char(36)" json:"modelID" yaml:"modelID"`
}

func (*ClientModelRelation) TableName() string { return "ai_proxy_client_model_relation" }

func (c *ClientModelRelation) ToProtobuf() *pb.ClientModelRelation {
	return &pb.ClientModelRelation{
		Id:        c.ID.String,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
		DeletedAt: timestamppb.New(c.DeletedAt.Time),
		ClientId:  c.ClientID,
		ModelId:   c.ModelID,
	}
}

type ClientModelRelations []*ClientModelRelation

func (c *ClientModelRelations) ToProtobuf() []*pb.ClientModelRelation {
	var pbRelations []*pb.ClientModelRelation
	for _, relation := range *c {
		pbRelations = append(pbRelations, relation.ToProtobuf())
	}
	return pbRelations
}
