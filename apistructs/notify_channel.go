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

package apistructs

// NotifyChannelFetchResponse 通知渠道详情响应结构
type NotifyChannelFetchResponse struct {
	Header
	Data NotifyChannelDTO `json:"data"`
}

// NotifyChannelDTO 通知渠道结构
type NotifyChannelDTO struct {
	ID     uint64               `json:"id"`
	Name   string               `json:"name"`
	Config *NotifyChannelConfig `json:"config"`
}

type NotifyChannelConfig struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	SignName        string `json:"signName"`
	TemplateCode    string `json:"templateCode"`
}
