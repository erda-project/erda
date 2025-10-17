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

	auditpb "github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

type Audit struct {
	common.BaseModel
	RequestAt  time.Time `gorm:"column:request_at;type:datetime" json:"requestAt" yaml:"requestAt"`
	ResponseAt time.Time `gorm:"column:response_at;type:datetime" json:"responseAt" yaml:"responseAt"`

	AuthKey            string `gorm:"column:auth_key;type:varchar(191)" json:"authKey" yaml:"authKey"`
	Status             int32  `gorm:"column:status;type:int(11)" json:"status" yaml:"status"`
	Prompt             string `gorm:"column:prompt;type:mediumtext" json:"prompt" yaml:"prompt"`
	Completion         string `gorm:"column:completion;type:longtext" json:"completion" yaml:"completion"`
	RequestBody        string `gorm:"column:request_body;type:longtext" json:"requestBody" yaml:"requestBody"`
	ResponseBody       string `gorm:"column:response_body;type:longtext" json:"responseBody" yaml:"responseBody"`
	ActualRequestBody  string `gorm:"column:actual_request_body;type:longtext" json:"actualRequestBody" yaml:"actualRequestBody"`
	ActualResponseBody string `gorm:"column:actual_response_body;type:longtext" json:"actualResponseBody" yaml:"actualResponseBody"`
	UserAgent          string `gorm:"column:user_agent;type:text" json:"userAgent" yaml:"userAgent"`
	XRequestID         string `gorm:"column:x_request_id;type:varchar(64)" json:"xRequestID" yaml:"xRequestID"`
	CallID             string `gorm:"column:call_id;type:char(36)" json:"callID" yaml:"callID"`

	ClientID  string `gorm:"column:client_id;type:char(36)" json:"clientID" yaml:"clientID"`
	ModelID   string `gorm:"column:model_id;type:char(36)" json:"modelID" yaml:"modelID"`
	SessionID string `gorm:"column:session_id;type:char(36)" json:"sessionID" yaml:"sessionID"`

	Username string `gorm:"column:username;type:varchar(128)" json:"username" yaml:"username"`
	Email    string `gorm:"column:email;type:varchar(64)" json:"email" yaml:"email"`

	BizSource   string `gorm:"column:source;type:varchar(128)" json:"bizSource" yaml:"bizSource"`
	OperationID string `gorm:"column:operation_id;type:varchar(128)" json:"operationID" yaml:"operationID"`

	ResponseFunctionCallName string `gorm:"column:res_func_call_name;type:varchar(128)" json:"responseFunctionCallName" yaml:"responseFunctionCallName"`

	Metadata metadata.Metadata `gorm:"column:metadata;type:mediumtext" json:"metadata" yaml:"metadata"`
}

func (*Audit) TableName() string { return "ai_proxy_filter_audit" }

func (audit *Audit) ToProtobuf() *auditpb.Audit {
	return &auditpb.Audit{
		Id:                       audit.ID.String,
		CreatedAt:                timestamppb.New(audit.CreatedAt),
		UpdatedAt:                timestamppb.New(audit.UpdatedAt),
		DeletedAt:                timestamppb.New(audit.DeletedAt.Time),
		RequestAt:                timestamppb.New(audit.RequestAt),
		ResponseAt:               timestamppb.New(audit.ResponseAt),
		AuthKey:                  audit.AuthKey,
		Status:                   audit.Status,
		Prompt:                   audit.Prompt,
		Completion:               audit.Completion,
		RequestBody:              audit.RequestBody,
		ResponseBody:             audit.ResponseBody,
		ActualRequestBody:        audit.ActualRequestBody,
		ActualResponseBody:       audit.ActualResponseBody,
		UserAgent:                audit.UserAgent,
		XRequestId:               audit.XRequestID,
		CallId:                   audit.CallID,
		ClientId:                 audit.ClientID,
		ModelId:                  audit.ModelID,
		SessionId:                audit.SessionID,
		Username:                 audit.Username,
		Email:                    audit.Email,
		BizSource:                audit.BizSource,
		OperationId:              audit.OperationID,
		ResponseFunctionCallName: audit.ResponseFunctionCallName,
		Metadata:                 audit.Metadata.ToProtobuf(),
	}
}

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

func (audits Audits) ToProtobuf() []*auditpb.Audit {
	var pbAudits []*auditpb.Audit
	for _, c := range audits {
		pbAudits = append(pbAudits, c.ToProtobuf())
	}
	return pbAudits
}
