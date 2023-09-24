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
	"google.golang.org/protobuf/types/known/timestamppb"

	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
)

func (this *AIProxyFilterAudit) ToChatLogProtobuf() *sessionpb.ChatLog {
	return &sessionpb.ChatLog{
		Id:         this.ID.String,
		RequestAt:  timestamppb.New(this.RequestAt),
		ResponseAt: timestamppb.New(this.ResponseAt),
		Prompt:     this.Prompt,
		Completion: this.Completion,
		SessionId:  this.SessionID,
	}
}

func (list *AIProxyFilterAuditList) ToChatLogsProtobuf() []*sessionpb.ChatLog {
	var chatLogs []*sessionpb.ChatLog
	for _, audit := range *list {
		chatLogs = append(chatLogs, audit.ToChatLogProtobuf())
	}
	return chatLogs
}
