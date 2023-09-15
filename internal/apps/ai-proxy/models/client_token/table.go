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

package client_token

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

type ClientToken struct {
	common.BaseModel
	ClientID  string            `gorm:"column:client_id;type:char(36)" json:"clientID" yaml:"clientID"`
	UserID    string            `gorm:"column:user_id;type:varchar(191)" json:"userID" yaml:"userID"`
	Token     string            `gorm:"column:token;type:char(34)" json:"token" yaml:"token"`
	ExpiredAt time.Time         `gorm:"column:expired_at;type:datetime" json:"expiredAt" yaml:"expiredAt"`
	Metadata  metadata.Metadata `gorm:"column:metadata;type:json" json:"metadata" yaml:"metadata"`
}

func (*ClientToken) TableName() string { return "ai_proxy_client_token" }

func (c *ClientToken) ToProtobuf() *pb.ClientToken {
	return &pb.ClientToken{
		Id:        c.ID.String,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
		DeletedAt: timestamppb.New(c.DeletedAt.Time),
		ClientId:  c.ClientID,
		UserId:    c.UserID,
		Token:     c.Token,
		ExpireAt:  timestamppb.New(c.ExpiredAt),
		Metadata:  c.Metadata.ToProtobuf(),
	}
}

type ClientTokens []*ClientToken

func (tokens ClientTokens) ToProtobuf() []*pb.ClientToken {
	var pbTokens []*pb.ClientToken
	for _, c := range tokens {
		pbTokens = append(pbTokens, c.ToProtobuf())
	}
	return pbTokens
}
