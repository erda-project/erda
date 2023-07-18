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

// AIProxyFilterAudit is the table ai_audit
type AIProxyFilterAudit struct {
	Id        fields.UUID      `json:"id" yaml:"id" gorm:"id"`
	CreatedAt time.Time        `json:"createdAt" yaml:"createdAt" gorm:"created_at"`
	UpdatedAt time.Time        `json:"updatedAt" yaml:"updatedAt" gorm:"updated_at"`
	DeletedAt fields.DeletedAt `json:"deletedAt" yaml:"deletedAt" gorm:"deleted_at"`

	// APIKeySha256 is the checksum of api-key with algorithm SHA256
	APIKeySha256 string `json:"apiKeySha256" yaml:"apiKeySha256" gorm:"api_key_sha256"`

	Username        string `json:"username" yaml:"username" gorm:"username"`
	PhoneNumber     string `json:"phoneNumber" yaml:"phoneNumber" gorm:"phone_number"`
	Email           string `json:"email" yaml:"email" gorm:"email"`
	DingtalkStaffId string `json:"dingtalkStaffId" yaml:"dingtalkStaffId" gorm:"dingtalk_staff_id"`
	JobNumber       string `json:"jobNumber" yaml:"jobNumber" yaml:"job_number"`

	// SessionId records the uniqueness of the conversation
	SessionId string `json:"sessionId" yaml:"sessionId" gorm:"session_id"`
	ChatType  string `json:"chatType" yaml:"chatType" gorm:"chat_type"`
	ChatTitle string `json:"chatTitle" yaml:"chatTitle" gorm:"chat_title"`
	ChatId    string `json:"chatId" yaml:"chatId" gorm:"chat_id"`
	// Source is the application source, like dingtalk, webui, vscode-plugin, jetbrains-plugin
	Source string `json:"source" yaml:"source" gorm:"source"`
	// Provider is an AI capability provider, like openai:chatgpt/v1, baidu:wenxin, alibaba:tongyi
	Provider string `json:"provider" yaml:"provider" gorm:"provider"`
	// Model used for this request, e.g. gpt-3.5-turbo, gpt-4-8k
	Model string `json:"model" yaml:"model" gorm:"model"`
	// OperationId is the unique identifier of the API
	OperationId string `json:"operationId" yaml:"operationId" gorm:"operation_id"`
	// Prompt The prompt(s) to generate completions for, encoded as a string, array of strings, array of tokens, or array of token arrays.
	//
	// Note that <|endoftext|> is the document separator that the model sees during training,
	// so if a prompt is not specified the model will generate as if from the beginning of a new document.
	Prompt string `json:"prompt" yaml:"prompt" gorm:"prompt"`
	// Completion returns the response to the client
	Completion string `json:"completion" yaml:"completion" gorm:"completion"`

	// XRequestId is the value of request header X-Request-Id
	XRequestId string `json:"XRequestId" yaml:"XRequestId" gorm:"x_request_id"`
	// RequestAt is the request arrival time
	RequestAt time.Time `json:"requestAt" yaml:"requestAt" gorm:"request_at"`
	// ResponseAt is the response arrival time
	ResponseAt time.Time `json:"responseAt" yaml:"responseAt" gorm:"response_at"`
	// UserAgent http client's User-Agent
	UserAgent           string `json:"userAgent" yaml:"userAgent" gorm:"user_agent"`
	RequestContentType  string `json:"requestContentType" yaml:"requestContentType" gorm:"request_content_type"`
	RequestBody         string `json:"requestBody" yaml:"requestBody" gorm:"request_body"`
	ResponseContentType string `json:"responseContentType" yaml:"responseContentType" gorm:"response_content_type"`
	ResponseBody        string `json:"responseBody" yaml:"responseBody" gorm:"response_body"`
	Server              string `json:"server" yaml:"server" gorm:"server"`
	Status              string `json:"status" yaml:"status" gorm:"status"`
	StatusCode          int    `json:"statusCode" yaml:"statusCode" gorm:"status_code"`
}

func (*AIProxyFilterAudit) TableName() string {
	return "ai_proxy_filter_audit"
}

func (audit *AIProxyFilterAudit) ToProtobufChatLog() *pb.ChatLog {
	return &pb.ChatLog{
		Id:         audit.Id.String,
		RequestAt:  timestamppb.New(audit.RequestAt),
		Prompt:     audit.Prompt,
		ResponseAt: timestamppb.New(audit.ResponseAt),
		Completion: audit.Completion,
	}
}

func (audit *AIProxyFilterAudit) SetAPIKeySha256(apiKey string) {
	audit.APIKeySha256 = Sha256(apiKey)
}

func Sha256(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}
