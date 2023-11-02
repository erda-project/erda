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
	AliyunBailianProviderMeta struct {
		Public AliyunBailianProviderMetaPublic `json:"public,omitempty"`
		Secret AliyunBailianProviderMetaSecret `json:"secret,omitempty"`
	}
	AliyunBailianProviderMetaPublic struct{}
	AliyunBailianProviderMetaSecret struct {
		AccessKeyId     string `json:"accessKeyId,omitempty"`
		AccessKeySecret string `json:"accessKeySecret,omitempty"`
		AgentKey        string `json:"agentKey,omitempty"`
	}
)

type (
	AliyunBailianModelMeta struct {
		Public AliyunBailianModelMetaPublic `json:"public,omitempty"`
		Secret AliyunBailianModelMetaSecret `json:"secret,omitempty"`
	}
	AliyunBailianModelMetaPublic struct{}
	AliyunBailianModelMetaSecret struct {
		AppId string `json:"appId,omitempty"`
	}
)
