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

package session

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

type Session struct {
	common.BaseModel

	ClientID string `gorm:"column:client_id;type:char(36)" json:"clientID" yaml:"clientID"`
	PromptID string `gorm:"column:prompt_id;type:char(36)" json:"promptID" yaml:"promptID"`
	ModelID  string `gorm:"column:model_id;type:char(36)" json:"modelID" yaml:"modelID"`
	Scene    string `gorm:"column:scene;type:varchar(128)" json:"scene" yaml:"scene"`
	UserID   string `gorm:"column:user_id;type:varchar(128)" json:"userID" yaml:"userID"`

	Name        string           `gorm:"column:name;type:varchar(128)" json:"name" yaml:"name"`
	Topic       string           `gorm:"column:topic;type:text" json:"topic" yaml:"topic"`
	NumOfCtxMsg int64            `gorm:"column:num_of_ctx_msg;type:int(11)" json:"numOfCtxMsg" yaml:"numOfCtxMsg"`
	IsArchived  bool             `gorm:"column:is_archived;type:tinyint(1)" json:"isArchived" yaml:"isArchived"`
	ResetAt     fields.DeletedAt `gorm:"column:reset_at;type:datetime" json:"resetAt" yaml:"resetAt"`
	Temperature float64          `gorm:"column:temperature;type:decimal(11,0)" json:"temperature" yaml:"temperature"`

	Metadata metadata.Metadata `gorm:"column:metadata;type:mediumtext" json:"metadata" yaml:"metadata"`
}

func (*Session) TableName() string { return "ai_proxy_session" }

func (s *Session) ToProtobuf() *pb.Session {
	return &pb.Session{
		Id:          s.ID.String,
		CreatedAt:   timestamppb.New(s.CreatedAt),
		UpdatedAt:   timestamppb.New(s.UpdatedAt),
		DeletedAt:   timestamppb.New(s.DeletedAt.Time),
		ClientId:    s.ClientID,
		PromptId:    s.PromptID,
		ModelId:     s.ModelID,
		Scene:       s.Scene,
		UserId:      s.UserID,
		Name:        s.Name,
		Topic:       s.Topic,
		NumOfCtxMsg: s.NumOfCtxMsg,
		IsArchived:  s.IsArchived,
		ResetAt:     timestamppb.New(s.ResetAt.Time),
		Temperature: s.Temperature,
		Metadata:    s.Metadata.ToProtobuf(),
	}
}
