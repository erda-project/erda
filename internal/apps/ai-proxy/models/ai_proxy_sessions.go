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
	"database/sql"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
)

// AIProxySessions is the table ai_proxy_sessions
type AIProxySessions struct {
	Id        fields.UUID      `json:"id" yaml:"id" gorm:"id"`
	CreatedAt time.Time        `json:"createdAt" yaml:"createdAt" gorm:"created_at"`
	UpdatedAt time.Time        `json:"updatedAt" yaml:"updatedAt" gorm:"updated_at"`
	DeletedAt fields.DeletedAt `json:"deletedAt" yaml:"deletedAt" gorm:"deleted_at"`

	UserID        string       `gorm:"type:varchar(128);not null;default:'';comment:用户id"`
	Name          string       `gorm:"type:varchar(128);not null;default:'';comment:会话名称"`
	Topic         string       `gorm:"type:text;not null;comment:会话主题"`
	ContextLength uint32       `gorm:"not null;default:0;comment:上下文长度"`
	Source        string       `gorm:"type:varchar(128);not null;comment:接入应用: dingtalk, vscode-plugin, jetbrains-plugin ..."`
	IsArchived    bool         `gorm:"not null;default:false;comment:是否归档"`
	ResetAt       sql.NullTime `gorm:"not null;default:'1970-01-01 00:00:00';comment:删除时间, 1970-01-01 00:00:00 表示未删除"`
	Model         string       `gorm:"type:varchar(128);not null;comment:调用的模型名称: gpt-3.5-turbo, gpt-4-8k, ..."`
	Temperature   float64      `gorm:"not null;default:0.7;comment:Higher values will make the output more random, while lower values will make it more focused and deterministic"`
}

func (*AIProxySessions) TableName() string {
	return "ai_proxy_sessions"
}

func (session *AIProxySessions) ToProtobuf() *pb.Session {
	return &pb.Session{
		Id:            session.Id.String,
		UserId:        session.UserID,
		Name:          session.Name,
		Topic:         session.Topic,
		ContextLength: session.ContextLength,
		IsArchived:    session.IsArchived,
		Source:        session.Source,
		ResetAt:       timestamppb.New(session.ResetAt.Time),
		Model:         session.Model,
		Temperature:   session.Temperature,
		CreatedAt:     timestamppb.New(session.CreatedAt),
		UpdatedAt:     timestamppb.New(session.UpdatedAt),
	}
}

func (session *AIProxySessions) WhereUserID() WhereField {
	return whereField{fieldName: "user_id"}
}

func (session *AIProxySessions) WhereSource() WhereField {
	return whereField{fieldName: "source"}
}
