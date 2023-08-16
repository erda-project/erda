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

package main

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"

	"github.com/erda-project/erda/apistructs"
)

var updateAutotestAPIStep = openai.FunctionDefinition{
	Name: "update-autotest-API-step",
	Description: `
必须遵循以下要求：
1. 以 Swagger 定义为准；如果 Swagger 里要求某个字段必填而用户没有说明，则自动生成一个有意义的不为空的随机值
2. 注意上下文变量的使用方式为：'${{ xxx }}'，注意花括号里两边的空格必填
`,
	Parameters: &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"params": {
				Type:        jsonschema.Array,
				Description: "HTTP Query Params",
				Items: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{ // see: apistructs.APIParam
						"key": {
							Type:        jsonschema.String,
							Description: "query param key",
						},
						"value": {
							Type:        jsonschema.String,
							Description: "query param value",
						},
						"desc": {
							Type:        jsonschema.String,
							Description: "query param description",
						},
					},
					Required: []string{"key", "value"},
				},
			},
			"headers": {
				Type:        jsonschema.Array,
				Description: "HTTP Headers",
				Items: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{ // see: apistructs.APIHeader
						"key": {
							Type:        jsonschema.String,
							Description: "header key",
						},
						"value": {
							Type:        jsonschema.String,
							Description: "header value",
						},
						"desc": {
							Type:        jsonschema.String,
							Description: "header description",
						},
					},
					Required: []string{"key", "value", "desc"},
				},
			},
			"body": {
				Type:        jsonschema.Object,
				Description: "HTTP Request Body",
				Properties: map[string]jsonschema.Definition{ // see: apistructs.APIBody
					"type": {
						Type: jsonschema.String,
						Enum: []string{
							apistructs.APIBodyTypeNone.String(),
							apistructs.APIBodyTypeText.String(),
							apistructs.APIBodyTypeTextPlain.String(),
							apistructs.APIBodyTypeApplicationJSON.String(),
							apistructs.APIBodyTypeApplicationJSON2.String(),
							apistructs.APIBodyTypeApplicationXWWWFormUrlencoded.String(),
						},
					},
					"content": {
						Type:        jsonschema.String,
						Description: "根据 API 的 Swagger 定义，生成请求体的内容",
					},
				},
				Required: []string{"type", "content"},
			},
			"out_params": {
				Type: jsonschema.Array,
				Description: `
对于出参，你需要先从用户输入中准确判断一共需要几个出参，有时候出参会隐含在断言的描述中。
API 接口调用后的出参，从响应体中使用表达式提取，用于后续断言。
expression 字段里不能使用上下文变量，只能使用 jq 语法。
参考以下示例，从 HTTP Response 中提取相应内容作为出参。
当 source=body:json 时，expression 为 jq 表达式，例如 '.name' 获取 json response 最外层的 name 字段;
当 source=header 时，expression 为 header key，例如 'X-Request-Id';
当 source=status 时，expression 为 'status' 字符串。
`,
				Items: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{ // see: apistructs.APIOutParam
						"key": {
							Type:        jsonschema.String,
							Description: "key of OutParam",
						},
						"source": {
							Type: jsonschema.String,
							Enum: []string{
								apistructs.APIOutParamSourceStatus.String(),
								apistructs.APIOutParamSourceBodyJson.String(),
								apistructs.APIOutParamSourceHeader.String(),
							},
							Description: "source of OutParam",
						},
						"expression": {
							Type: jsonschema.String,
						},
					},
				},
				Required: []string{"key", "source", "expression"},
			},
			"asserts": {
				Type: jsonschema.Array,
				Description: `
Assertion of the output parameters after API interface invocation.
`,
				Items: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{ // see: apistructs.APIAssert
						"arg": {
							Type:        jsonschema.String,
							Description: "The parameters of the assertion can only be selected from the keys of the output parameters above.",
						},
						"operator": { // see: pkg/assert
							Type: jsonschema.String,
							Enum: []string{"=", "!=", ">=", "<=", ">", "<", "contains", "not_contains", "belong", "not_belong", "empty", "not_empty", "exist", "not_exist"},
						},
						"value": {
							Type:        jsonschema.String,
							Description: "Expected assertion result",
						},
					},
					Required: []string{"arg", "operator"},
				},
			},
		},
		Required: []string{"params", "headers", "body", "out_params", "asserts"},
	},
}
