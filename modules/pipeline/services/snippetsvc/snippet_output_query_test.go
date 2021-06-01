// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package snippetsvc

import (
	"encoding/json"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/expression"
)

func Test_getActionDetail(t *testing.T) {
	var table = []struct {
		config  apistructs.SnippetDetailQuery
		spec    string
		outputs []string
	}{
		{
			config: apistructs.SnippetDetailQuery{
				Alias: "jsonparse",
				SnippetConfig: apistructs.SnippetConfig{
					Name:   "jsonparse",
					Source: apistructs.ActionSourceType,
					Labels: map[string]string{
						"actionJson":    "{\"alias\":\"jsonparse\",\"type\":\"jsonparse\",\"description\":\"对 json 解析值然后返回出参\",\"version\":\"1.0\",\"params\":{\"data\":\"{\\\"aaa\\\": 1}\",\"out_params\":[{\"expression\":\".aaa\",\"key\":\"aaa\"}]},\"resources\":{},\"displayName\":\"json 解析\"}",
						"actionVersion": "1.0",
					},
				},
			},
			outputs: []string{
				"name",
				"name1",
				"key1",
			},
			spec: "name: jsonparse\nversion: '1.0'\ntype: action\ncategory: custom_task\ndisplayName: json 解析\ndesc: 对 json 解析值然后返回出参\npublic: true\nsupportedVersions:\n  - \">= 3.21\"\nlabels:\n  configsheet: true\n\nparams:\n  - name: out_params\n    required: false\n    desc: 出参\n    type: struct_array\n    struct:\n      - name: key\n        required: true\n        desc: 出参名\n      - name: expression\n        required: true\n        desc: 支持 linux jq 表达式 以 . 开头，支持 jackson 以 $. 开头\n  - name: data\n    required: true\n    desc: json 数据\n    \n    \noutputsFromParams:\n  - type: jq\n    keyExpr: \"[.out_params[].key]\"    \n\n",
		},
		{
			config: apistructs.SnippetDetailQuery{
				Alias: "jsonparse",
				SnippetConfig: apistructs.SnippetConfig{
					Name:   "jsonparse",
					Source: apistructs.ActionSourceType,
					Labels: map[string]string{
						"actionJson":    "{\"alias\":\"jsonparse\",\"type\":\"jsonparse\",\"description\":\"对 json 解析值然后返回出参\",\"version\":\"1.0\",\"params\":{\"data\":\"{\\\"aaa\\\": 1}\",\"out_params\":[{\"expression\":\".aaa\",\"key\":\"aaa\"}]},\"resources\":{},\"displayName\":\"json 解析\"}",
						"actionVersion": "1.0",
					},
				},
			},
			spec: "name: jsonparse\nversion: '1.0'\ntype: action\ncategory: custom_task\ndisplayName: json 解析\ndesc: 对 json 解析值然后返回出参\npublic: true\nsupportedVersions:\n  - \">= 3.21\"\nlabels:\n  configsheet: true\n\nparams:\n  - name: out_params\n    required: false\n    desc: 出参\n    type: struct_array\n    struct:\n      - name: key\n        required: true\n        desc: 出参名\n      - name: expression\n        required: true\n        desc: 支持 linux jq 表达式 以 . 开头，支持 jackson 以 $. 开头\n  - name: data\n    required: true\n    desc: json 数据\n    \n    \noutputsFromParams:\n  - type: jq\n    keyExpr: \"[.out_params[].key]\"    \n\n",
		},
		{
			config: apistructs.SnippetDetailQuery{
				Alias: "jsonparse",
				SnippetConfig: apistructs.SnippetConfig{
					Name:   "jsonparse",
					Source: apistructs.ActionSourceType,
					Labels: map[string]string{
						"actionJson":    "{\"alias\":\"jsonparse\",\"type\":\"jsonparse\",\"description\":\"对 json 解析值然后返回出参\",\"version\":\"1.0\",\"params\":{\"data\":\"{\\\"aaa\\\": 1}\",\"out_params\":[{\"expression\":\".aaa\",\"key\":\"aaa\"}]},\"resources\":{},\"displayName\":\"json 解析\"}",
						"actionVersion": "1.0",
					},
				},
			},
			outputs: []string{
				"result",
			},
			spec: "name: jsonparse\nversion: '1.0'\ntype: action\ncategory: custom_task\ndisplayName: json 解析\ndesc: 对 json 解析值然后返回出参\npublic: true\nsupportedVersions:\n  - \">= 3.21\"\nlabels:\n  configsheet: true\n\nparams:\n  - name: out_params\n    required: false\n    desc: 出参\n    type: struct_array\n    struct:\n      - name: key\n        required: true\n        desc: 出参名\n      - name: expression\n        required: true\n        desc: 支持 linux jq 表达式 以 . 开头，支持 jackson 以 $. 开头\n  - name: data\n    required: true\n    desc: json 数据\n\n\noutputs:\n  - name: result",
		},
	}

	var s = &SnippetSvc{}

	for _, data := range table {

		bdl := &bundle.Bundle{}
		s.bdl = bdl

		monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetExtensionVersion", func(b *bundle.Bundle, req apistructs.ExtensionVersionGetRequest) (*apistructs.ExtensionVersion, error) {
			return &apistructs.ExtensionVersion{
				Spec: data.spec,
			}, nil
		})

		monkey.Patch(handlerActionOutputsWithJq, func(action *apistructs.PipelineYmlAction, jq string) ([]string, error) {
			return data.outputs, nil
		})

		detail, err := s.getActionDetail(data.config)
		assert.NoError(t, err)

		assert.Equal(t, len(data.outputs), len(data.outputs))
		for _, key := range detail.Outputs {
			var find = false
			for _, output := range data.outputs {
				if key == expression.GenOutputRef(data.config.Alias, output) {
					find = true
				}
			}
			assert.True(t, find)
		}
	}
}

func Test_ActionJson(t *testing.T) {

	var str = "{\"alias\":\"account-login\",\"type\":\"api-test\",\"description\":\"执行单个接口测试。上层可以通过 pipeline.yml 编排一组接口测试的执行顺序。\",\"version\":\"2.0\",\"params\":{\"asserts\":[{\"arg\":\"status\",\"operator\":\"=\",\"value\":\"200\"},{\"arg\":\"sessionId\",\"operator\":\"not_empty\"}],\"headers\":[{\"key\":\"Content-Type\",\"value\":\"application/x-www-form-urlencoded\"}],\"method\":\"POST\",\"name\":\"登录\",\"out_params\":[{\"expression\":\".sessionid\",\"key\":\"sessionId\",\"source\":\"body:json\"},{\"key\":\"status\",\"source\":\"status\"},{\"expression\":\".id\",\"key\":\"userId\",\"source\":\"body:json\"}],\"params\":[{\"key\":\"username\",\"value\":\"${params.username}\"},{\"key\":\"password\",\"value\":\"${params.password}\"}],\"url\":\"${params.openapi_addr}/login\"},\"resources\":{},\"displayName\":\"接口测试\",\"logoUrl\":\"//terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/10/10/24195384-07b7-4203-93e1-666373639af4.png\"}"

	var action apistructs.PipelineYmlAction
	err := json.Unmarshal([]byte(str), &action)
	if err != nil {
		t.Fail()
	}

	params := action.Params
	if params == nil {
		t.Fail()
	}

	outParamsBytes, err := json.Marshal(action.Params["out_params"])
	if err != nil {
		t.Fail()
	}

	var outParams []apistructs.APIOutParam
	err = json.Unmarshal(outParamsBytes, &outParams)
	if err != nil {
		t.Fail()
	}

}
