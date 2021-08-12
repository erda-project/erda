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
