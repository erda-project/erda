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

package api_style

// APIStyle defines the style of the API used when invoking the concrete model.
// Currently, only provider-level API-style is supported.
type APIStyle string

const (
	// see: https://platform.openai.com/docs/api-reference/chat
	APIStyleOpenAICompatible APIStyle = "OpenAI-Compatible"

	// see: https://help.aliyun.com/zh/model-studio/use-qwen-by-calling-api
	APIStyleAliyunDashScope APIStyle = "AliyunDashScope"

	// see: https://docs.anthropic.com/en/api/messages
	// see: https://docs.aws.amazon.com/zh_cn/bedrock/latest/APIReference/API_runtime_InvokeModel.html
	APIStyleAnthropicCompatible APIStyle = "Anthropic-Compatible"
)

func (s APIStyle) IsValid() bool {
	switch s {
	case APIStyleOpenAICompatible, APIStyleAliyunDashScope, APIStyleAnthropicCompatible:
		return true
	default:
		return false
	}
}

func AllAPIStyles() []APIStyle {
	return []APIStyle{
		APIStyleOpenAICompatible,
		APIStyleAliyunDashScope,
		APIStyleAnthropicCompatible,
	}
}

type APIStyleConfig struct {
	// Method for the API, e.g., POST, GET.
	// default is empty, means to use request method.
	Method string `json:"method,omitempty"` // e.g., POST, GET

	// e.g., https, http
	Scheme string `json:"scheme,omitempty"`

	// host for the API, e.g., api.openai.com; dashscope.aliyuncs.com
	Host string `json:"host,omitempty"`

	// path for the API, e.g.,
	// - openai: /v1/chat/completions;
	// - volcano: /api/v3/chat/completions;
	// - azure openai: /openai/deployments/${@model.metadata.public.deployment_id}/chat/completions
	// host should be []string format, support op, see: PathOp
	// - set, ${full-path}
	// - replace, ${old}, ${new}
	// To be compatible with old data (set string), use type: any and auto convert to [set, ${full-path}]
	Path any `json:"path,omitempty"`
	// Supported Path Op: set, replace
	PathOp []string `json:"-"`

	// Custom query parameters for the API.
	// the first element of a key is the operation type, support: "Add", "Set", "Delete".
	// ${@key||default-value} is supported, see: @JSONPathParser
	// e.g.,
	// - AzureOpenAI: api-version=${@model.metadata.public.api_version||@provider.metadata.public.api_version||2025-03-01-preview}
	//   -> api-version: []string{"Add", "2025-03-01-preview"}
	//   -> api-version: []string{"Delete"}
	QueryParams map[string][]string `json:"queryParams,omitempty"`

	// Custom headers for the API.
	// the first element of a key is the operation type, see above: @QueryParams
	// ${@key||default-value} is supported, see: @JSONPathParser
	// e.g.,
	// - AzureOpenAI:
	//   -> Authorization: []string{"Delete"}
	//   -> Api-Key: []string{"Add", "provider.api_key"}
	Headers map[string][]string `json:"headers,omitempty"`
}

// APIVendor based on the APIStyle, indicates the vendor of the API.
//
// For some reason, although the APIStyle is same, the body detail maybe a bit different, e.g.,
// - APIStyle: Anthropic-Compatible
//   - APIVendor: Anthropic vs AWS-Bedrock
//   - different: bedrock set model name in path, while Anthropic set model name in body
//
// So, this field is necessary to distinguish the implementation details inside one APIStyle director.
type APIVendor string
