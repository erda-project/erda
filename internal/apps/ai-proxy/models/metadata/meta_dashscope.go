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

import "fmt"

type AliyunDashScopeRequestType string

const (
	AliyunDashScopeRequestTypeOpenAI AliyunDashScopeRequestType = "openai"
	AliyunDashScopeRequestTypeDs     AliyunDashScopeRequestType = "ds"
)

func (t AliyunDashScopeRequestType) String() string {
	return string(t)
}
func (t AliyunDashScopeRequestType) Valid() (bool, error) {
	if t.String() == "" {
		return false, fmt.Errorf("empty request_type")
	}
	switch t {
	case AliyunDashScopeRequestTypeOpenAI, AliyunDashScopeRequestTypeDs:
		return true, nil
	default:
		return false, fmt.Errorf("unknown request_type: %s", t)
	}
}

type AliyunDashScopeResponseType string

const (
	AliyunDashScopeResponseTypeOpenAI AliyunDashScopeResponseType = "openai"
	AliyunDashScopeResponseTypeDs     AliyunDashScopeResponseType = "ds"
)

func (t AliyunDashScopeResponseType) String() string {
	return string(t)
}
func (t AliyunDashScopeResponseType) Valid() (bool, error) {
	if t.String() == "" {
		return false, fmt.Errorf("empty request_type")
	}
	switch t {
	case AliyunDashScopeResponseTypeOpenAI, AliyunDashScopeResponseTypeDs:
		return true, nil
	default:
		return false, fmt.Errorf("unknown request_type: %s", t)
	}
}

type (
	AliyunDashScopeModelMeta struct {
		Public AliyunDashScopeModelMetaPublic `json:"public,omitempty"`
		Secret AliyunDashScopeModelMetaSecret `json:"secret,omitempty"`
	}
	AliyunDashScopeModelMetaPublic struct {
		ModelName    string                      `json:"model_name,omitempty"`
		RequestType  AliyunDashScopeRequestType  `json:"request_type,omitempty"`
		ResponseType AliyunDashScopeResponseType `json:"response_type,omitempty"`
		CustomURL    string                      `json:"custom_url"` // e.g., https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions
	}
	AliyunDashScopeModelMetaSecret struct {
	}
)
