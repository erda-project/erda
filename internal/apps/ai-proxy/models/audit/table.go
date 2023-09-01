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

package audit

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
)

type Audit struct {
	common.BaseModel
	APIKeySHA256        string    `gorm:"column:api_key_sha256;type:char(64)" json:"aPIKeySHA256" yaml:"aPIKeySHA256"`
	Username            string    `gorm:"column:username;type:varchar(128)" json:"username" yaml:"username"`
	PhoneNumber         string    `gorm:"column:phone_number;type:varchar(32)" json:"phoneNumber" yaml:"phoneNumber"`
	JobNumber           string    `gorm:"column:job_number;type:varchar(32)" json:"jobNumber" yaml:"jobNumber"`
	Email               string    `gorm:"column:email;type:varchar(64)" json:"email" yaml:"email"`
	DingtalkStaffID     string    `gorm:"column:dingtalk_staff_id;type:varchar(64)" json:"dingtalkStaffID" yaml:"dingtalkStaffID"`
	SessionID           string    `gorm:"column:session_id;type:varchar(64)" json:"sessionID" yaml:"sessionID"`
	ChatType            string    `gorm:"column:chat_type;type:varchar(32)" json:"chatType" yaml:"chatType"`
	ChatTitle           string    `gorm:"column:chat_title;type:varchar(128)" json:"chatTitle" yaml:"chatTitle"`
	ChatID              string    `gorm:"column:chat_id;type:varchar(64)" json:"chatID" yaml:"chatID"`
	Source              string    `gorm:"column:source;type:varchar(128)" json:"source" yaml:"source"`
	ProviderName        string    `gorm:"column:provider_name;type:varchar(128)" json:"providerName" yaml:"providerName"`
	ProviderInstanceID  string    `gorm:"column:provider_instance_id;type:varchar(512)" json:"providerInstanceID" yaml:"providerInstanceID"`
	Model               string    `gorm:"column:model;type:varchar(128)" json:"model" yaml:"model"`
	OperationID         string    `gorm:"column:operation_id;type:varchar(128)" json:"operationID" yaml:"operationID"`
	Prompt              string    `gorm:"column:prompt;type:mediumtext" json:"prompt" yaml:"prompt"`
	Completion          string    `gorm:"column:completion;type:longtext" json:"completion" yaml:"completion"`
	ReqFuncCallName     string    `gorm:"column:req_func_call_name;type:varchar(128)" json:"reqFuncCallName" yaml:"reqFuncCallName"`
	ReqFuncCallArgs     string    `gorm:"column:req_func_call_args;type:longtext" json:"reqFuncCallArgs" yaml:"reqFuncCallArgs"`
	ResFuncCallName     string    `gorm:"column:res_func_call_name;type:varchar(128)" json:"resFuncCallName" yaml:"resFuncCallName"`
	ResFuncCallArgs     string    `gorm:"column:res_func_call_args;type:longtext" json:"resFuncCallArgs" yaml:"resFuncCallArgs"`
	Metadata            string    `gorm:"column:metadata;type:longtext" json:"metadata" yaml:"metadata"`
	XRequestID          string    `gorm:"column:x_request_id;type:varchar(64)" json:"xRequestID" yaml:"xRequestID"`
	RequestAt           time.Time `gorm:"column:request_at;type:datetime" json:"requestAt" yaml:"requestAt"`
	ResponseAt          time.Time `gorm:"column:response_at;type:datetime" json:"responseAt" yaml:"responseAt"`
	RequestContentType  string    `gorm:"column:request_content_type;type:varchar(32)" json:"requestContentType" yaml:"requestContentType"`
	RequestBody         string    `gorm:"column:request_body;type:longtext" json:"requestBody" yaml:"requestBody"`
	ResponseContentType string    `gorm:"column:response_content_type;type:varchar(32)" json:"responseContentType" yaml:"responseContentType"`
	ResponseBody        string    `gorm:"column:response_body;type:longtext" json:"responseBody" yaml:"responseBody"`
	UserAgent           string    `gorm:"column:user_agent;type:text" json:"userAgent" yaml:"userAgent"`
	Server              string    `gorm:"column:server;type:varchar(32)" json:"server" yaml:"server"`
	Status              string    `gorm:"column:status;type:varchar(32)" json:"status" yaml:"status"`
	StatusCode          int64     `gorm:"column:status_code;type:int(11)" json:"statusCode" yaml:"statusCode"`
}

func (*Audit) TableName() string { return "ai_proxy_filter_audit" }

func (audit *Audit) ToChatLogProtobuf() *sessionpb.ChatLog {
	return &sessionpb.ChatLog{
		Id:         audit.ID.String,
		RequestAt:  timestamppb.New(audit.RequestAt),
		ResponseAt: timestamppb.New(audit.ResponseAt),
		Prompt:     audit.Prompt,
		Completion: audit.Completion,
		SessionId:  audit.SessionID,
	}
}

type Audits []*Audit

func (audits *Audits) ToChatLogsProtobuf() []*sessionpb.ChatLog {
	var chatLogs []*sessionpb.ChatLog
	for _, audit := range *audits {
		chatLogs = append(chatLogs, audit.ToChatLogProtobuf())
	}
	return chatLogs
}
