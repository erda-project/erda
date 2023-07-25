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
	"crypto/sha256"
	"encoding/hex"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
)

type AIProxyFilterAudit struct {
	ID        fields.UUID      `gorm:"type:char(36);primaryKey;comment:primary key"`
	CreatedAt time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP;comment:更新时间"`
	DeletedAt fields.DeletedAt `gorm:"not null;default:'1970-01-01 00:00:00';comment:删除时间, 1970-01-01 00:00:00 表示未删除"`

	APIKeySHA256        string    `gorm:"type:char(64);not null;default:'';comment:请求使用的 app_key sha256 哈希值"`
	Username            string    `gorm:"type:varchar(128);not null;comment:用户名称, source=dingtalk时, 为钉钉用户名称"`
	PhoneNumber         string    `gorm:"type:varchar(32);not null;comment:用户手机号码, source=dingtalk时, 为钉钉账号注册手机号"`
	JobNumber           string    `gorm:"type:varchar(32);not null;comment:用户工号, source=dingtalk时, 为用户在其组织内的工号"`
	Email               string    `gorm:"type:varchar(64);not null;comment:用户邮箱"`
	DingTalkStaffID     string    `gorm:"column:dingtalk_staff_id;type:varchar(64);not null;comment:用户钉钉号"`
	SessionID           string    `gorm:"type:varchar(64);not null;comment:对话标识"`
	ChatType            string    `gorm:"type:varchar(32);not null;comment:对话类型"`
	ChatTitle           string    `gorm:"type:varchar(128);not null;comment:source=dingtalk时, 私聊时为 private, 群聊时为群名称"`
	ChatID              string    `gorm:"type:varchar(64);not null;comment:钉钉聊天 id"`
	Source              string    `gorm:"type:varchar(128);not null;comment:接入应用: dingtalk, vscode-plugin, jetbrains-plugin ..."`
	ProviderName        string    `gorm:"type:varchar(128);not null;comment:AI 能力提供商: openai, azure..."`
	ProviderInstanceID  string    `gorm:"type:varchar(512);not null;default:'';comment:provider 实例 id"`
	Model               string    `gorm:"type:varchar(128);not null;comment:调用的模型名称: gpt-3.5-turbo, gpt-4-8k, ..."`
	OperationID         string    `gorm:"type:varchar(128);not null;comment:调用的接口名称, HTTP Method + Path"`
	Prompt              string    `gorm:"type:mediumtext;not null;comment:提示语"`
	Completion          string    `gorm:"type:longtext;not null;comment:AI 回复多个 choices 中的一个"`
	Metadata            string    `gorm:"type:longtext;not null;comment:客户端要审计的其他信息"`
	XRequestID          string    `gorm:"type:varchar(64);not null;default:'';comment:http 请求中的 X-Request-Id"`
	RequestAt           time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;comment:请求到达时间"`
	ResponseAt          time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;comment:响应到达时间"`
	RequestContentType  string    `gorm:"type:varchar(32);not null;comment:请求使用的 Content-Type"`
	RequestBody         string    `gorm:"type:longtext;not null;comment:请求的 Body"`
	ResponseContentType string    `gorm:"type:varchar(32);not null;comment:响应使用的 Content-Type"`
	ResponseBody        string    `gorm:"type:longtext;not null;comment:响应的 Body"`
	UserAgent           string    `gorm:"type:varchar(128);not null;comment:http 客户端 User-Agent"`
	Server              string    `gorm:"type:varchar(32);not null;comment:response server"`
	Status              string    `gorm:"type:varchar(32);not null;comment:http response status"`
	StatusCode          int       `gorm:"not null;comment:http response status code"`
}

func (*AIProxyFilterAudit) TableName() string {
	return "ai_proxy_filter_audit"
}

func (audit *AIProxyFilterAudit) ToProtobufChatLog() *pb.ChatLog {
	return &pb.ChatLog{
		Id:         audit.ID.String,
		RequestAt:  timestamppb.New(audit.RequestAt),
		Prompt:     audit.Prompt,
		ResponseAt: timestamppb.New(audit.ResponseAt),
		Completion: audit.Completion,
	}
}

func (audit *AIProxyFilterAudit) SetAPIKeySha256(apiKey string) {
	audit.APIKeySHA256 = Sha256(apiKey)
}

func Sha256(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}
