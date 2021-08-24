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

package action

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func Test_fillTestPlanFields(t *testing.T) {
	type args struct {
		field     []apistructs.FormPropItem
		testPlans []map[string]interface{}
		cms       []map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want []apistructs.FormPropItem
	}{
		// TODO: Add test cases.
		{
			name: "Filled",
			args: args{
				field: []apistructs.FormPropItem{
					apistructs.FormPropItem{
						Label:     "执行条件",
						Component: "input",
						Required:  true,
						Group:     "params",
					},
				},
				testPlans: []map[string]interface{}{
					map[string]interface{}{
						"name":  "a",
						"value": "1",
					},
					map[string]interface{}{
						"name":  "b",
						"value": "2",
					},
				},
				cms: []map[string]interface{}{
					map[string]interface{}{
						"name":  "aa",
						"value": "11",
					},
					map[string]interface{}{
						"name":  "bb",
						"value": "22",
					},
				},
			},
			want: []apistructs.FormPropItem{
				apistructs.FormPropItem{
					Label:     "执行条件",
					Component: "input",
					Required:  true,
					Group:     "params",
				},
				apistructs.FormPropItem{
					Component: "formGroup",
					ComponentProps: map[string]interface{}{
						"title": "任务参数",
					},
					Group: "params",
					Key:   "params",
				},
				apistructs.FormPropItem{
					Label:     "测试计划",
					Component: "select",
					Required:  true,
					Key:       "params.test_plan",
					ComponentProps: map[string]interface{}{
						"options": []map[string]interface{}{
							map[string]interface{}{
								"name":  "a",
								"value": 1,
							},
							map[string]interface{}{
								"name":  "b",
								"value": 2,
							},
						},
					},
					Group: "params",
				},
				apistructs.FormPropItem{
					Label:     "参数配置",
					Component: "select",
					Required:  true,
					Key:       "params.cms",
					ComponentProps: map[string]interface{}{
						"options": []map[string]interface{}{
							map[string]interface{}{
								"name":  "aa",
								"value": 11,
							},
							map[string]interface{}{
								"name":  "bb",
								"value": 22,
							},
						},
					},
					Group: "params",
				},
				apistructs.FormPropItem{
					Label:        "等待执行结果",
					Component:    "input",
					Required:     false,
					Key:          "params.waiting_result",
					Group:        "params",
					DefaultValue: false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fillTestPlanFields(tt.args.field, tt.args.testPlans, tt.args.cms); !reflect.DeepEqual(got, tt.want) {
				fmt.Println(got)
				fmt.Println(tt.want)
			}
		})
	}
}
