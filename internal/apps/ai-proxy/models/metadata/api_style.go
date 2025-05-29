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

// APIStyle defines the style of the API used when invoking the concrete model.
// Currently, only provider-level API-style is supported.
type APIStyle string

const (
	// see: https://platform.openai.com/docs/api-reference/chat
	APIStyleOpenAICompatible APIStyle = "OpenAI-Compatible"

	// see: https://help.aliyun.com/zh/model-studio/use-qwen-by-calling-api
	APIStyleAliyunDashScope APIStyle = "AliyunDashScope"
)

func (s APIStyle) IsValid() bool {
	switch s {
	case APIStyleOpenAICompatible, APIStyleAliyunDashScope:
		return true
	default:
		return false
	}
}

func AllAPIStyles() []APIStyle {
	return []APIStyle{
		APIStyleOpenAICompatible,
		APIStyleAliyunDashScope,
	}
}

type APIStyleConfig struct {
	// Method for the API, e.g., POST, GET.
	// default is POST.
	Method string `json:"method,omitempty"` // e.g., POST, GET

	// e.g., https, http
	Scheme string `json:"scheme,omitempty"`

	// host for the API, e.g., api.openai.com; dashscope.aliyuncs.com
	Host string `json:"host,omitempty"`

	// path for the API, e.g.,
	// - openai: /v1/chat/completions;
	// - volcano: /api/v3/chat/completions;
	// - azure openai: /openai/deployments/${model.metadata.public.extra.deployment_id}/chat/completions
	Path string `json:"path,omitempty"`

	// Custom query parameters for the API.
	// the first element of a key is the operation type, support: "Add", "Set", "Delete".
	// e.g.,
	// - AzureOpenAI: api-version=${model.metadata.public.extra.api_version||provider.metadata.public.extra.api_version||2025-03-01-preview}
	//   -> api-version: []string{"Add", "2025-03-01-preview"}
	//   -> api-version: []string{"Delete"}
	QueryParams map[string][]string `json:"queryParams,omitempty"`

	// Custom headers for the API.
	// the first element of a key is the operation type, see above: @QueryParams
	// e.g.,
	// - AzureOpenAI:
	//   -> Authorization: []string{"Delete"}
	//   -> Api-Key: []string{"Add", "provider.api_key"}
	Headers map[string][]string `json:"headers,omitempty"`
}
