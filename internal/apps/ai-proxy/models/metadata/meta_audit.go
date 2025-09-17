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

package metadata

type (
	AuditMetadata struct {
		Public AuditMetadataPublic `json:"public,omitempty"`
		Secret AuditMetadataSecret `json:"secret,omitempty"`
	}
	AuditMetadataPublic struct {
		FilterName  string `json:"filter.name,omitempty"`
		FilterError string `json:"filter.error,omitempty"`

		RequestContentType      string `json:"request.content_type,omitempty"`
		ActualRequestURL        string `json:"actual.request.url,omitempty"`
		RequestFunctionCallName string `json:"request.function_call.name,omitempty"`

		ResponseContentType      string `json:"response.content_type,omitempty"`
		ResponseFunctionCallName string `json:"response.function_call.name,omitempty"`

		TimeCost string `json:"time_cost,omitempty"`

		AudioFileName    string `json:"audio.file.name,omitempty"`
		AudioFileSize    string `json:"audio.file.size,omitempty"`
		AudioFileHeaders string `json:"audio.file.headers,omitempty"`

		ImageQuality string `json:"image.quality,omitempty"`
		ImageSize    string `json:"image.size,omitempty"`
		ImageStyle   string `json:"image.style,omitempty"`
	}
	AuditMetadataSecret struct {
		IdentityPhoneNumber string `json:"identity.phone_number,omitempty"`
		IdentityJobNumber   string `json:"identity.job_number,omitempty"`

		DingtalkStaffId   string `json:"dingtalk.staff_id,omitempty"`
		DingtalkChatType  string `json:"dingtalk.chat_type,omitempty"`
		DingtalkChatTitle string `json:"dingtalk.chat_title,omitempty"`
		DingtalkChatId    string `json:"dingtalk.chat_id,omitempty"`
	}
)
