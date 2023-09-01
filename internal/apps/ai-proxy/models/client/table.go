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

package client

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

type Client struct {
	common.BaseModel
	Name        string            `gorm:"column:name;type:varchar(191)" json:"name" yaml:"name"`
	Desc        string            `gorm:"column:desc;type:varchar(1024)" json:"desc" yaml:"desc"`
	AccessKeyID string            `gorm:"column:access_key_id;type:char(32)" json:"accessKeyID" yaml:"accessKeyID"`
	SecretKeyID string            `gorm:"column:secret_key_id;type:char(32)" json:"secretKeyID" yaml:"secretKeyID"`
	Metadata    metadata.Metadata `gorm:"column:metadata;type:mediumtext" json:"metadata" yaml:"metadata"`
}

func (*Client) TableName() string { return "ai_proxy_client" }

func (c *Client) ToProtobuf() *pb.Client {
	return &pb.Client{
		Id:          c.ID.String,
		CreatedAt:   timestamppb.New(c.CreatedAt),
		UpdatedAt:   timestamppb.New(c.UpdatedAt),
		DeletedAt:   timestamppb.New(c.DeletedAt.Time),
		Name:        c.Name,
		Desc:        c.Desc,
		AccessKeyId: c.AccessKeyID,
		SecretKeyId: c.SecretKeyID,
		Metadata:    c.Metadata.ToProtobuf(),
	}
}

type Clients []*Client

func (clients Clients) ToProtobuf() []*pb.Client {
	var pbClients []*pb.Client
	for _, c := range clients {
		pbClients = append(pbClients, c.ToProtobuf())
	}
	return pbClients
}
