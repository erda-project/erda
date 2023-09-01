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

package prompt

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

type Prompt struct {
	common.BaseModel
	Name     string            `gorm:"column:name;type:varchar(191)" json:"name" yaml:"name"`
	Desc     string            `gorm:"column:desc;type:varchar(1024)" json:"desc" yaml:"desc"`
	ClientID string            `gorm:"column:client_id;type:char(36)" json:"clientID" yaml:"clientID"`
	Messages message.Messages  `gorm:"column:messages;type:longtext" json:"messages" yaml:"messages"`
	Metadata metadata.Metadata `gorm:"column:metadata;type:mediumtext" json:"metadata" yaml:"metadata"`
}

func (*Prompt) TableName() string { return "ai_proxy_prompt" }

func (p *Prompt) ToProtobuf() *pb.Prompt {
	return &pb.Prompt{
		Id:        p.ID.String,
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
		DeletedAt: timestamppb.New(p.DeletedAt.Time),
		Name:      p.Name,
		Desc:      p.Desc,
		ClientId:  p.ClientID,
		Messages:  p.Messages.ToProtobuf(),
		Metadata:  p.Metadata.ToProtobuf(),
	}
}

type Prompts []Prompt

func (prompts *Prompts) ToProtobuf() []*pb.Prompt {
	var result []*pb.Prompt
	for _, prompt := range *prompts {
		result = append(result, prompt.ToProtobuf())
	}
	return result
}
