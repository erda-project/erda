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

package auto_test_plan_list

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// GenCreateFormModalProps 生成创建测试计划表单的props
func GenCreateFormModalProps(testSpace []byte) interface{} {
	props := `{
          "name": "计划",
          "fields": [
            {
              "component": "input",
              "key": "name",
              "label": "计划名称",
              "required": true,
              "rule": [
                {
                  "pattern": "/^[a-z\u4e00-\u9fa5A-Z0-9_-]*$/",
                  "msg": "可输入中文、英文、数字、中划线或下划线"
                }
              ],
              "componentProps": {
                "maxLength": 50
              }
            },
            {
              "component": "select",
              "key": "spaceId",
              "label": "测试空间",
							"disabled": false,
              "required": true,
              "componentProps": {
                "options": ` + string(testSpace) +
		`}
            },
            {
              "key": "owners",
              "label": "负责人",
              "required": true,
              "component": "memberSelector",
              "componentProps": {
                "mode": "multiple",
                "scopeType": "project"
              }
            }
          ]
        }`

	var propsI interface{}
	if err := json.Unmarshal([]byte(props), &propsI); err != nil {
		logrus.Errorf("init props name=testplan component=formModal propsType=CreateTestPlan err: errMsg: %v", err)
	}

	return propsI
}

// GenUpdateFormModalProps 生成更新测试计划表单的props
func GenUpdateFormModalProps(testSpace []byte) interface{} {
	props := `{
          "name": "计划",
          "fields": [
            {
              "component": "input",
              "key": "name",
              "label": "计划名称",
              "required": true,
              "rule": [
                {
                  "pattern": "/^[a-z\u4e00-\u9fa5A-Z0-9_-]*$/",
                  "msg": "可输入中文、英文、数字、中划线或下划线"
                }
              ],
              "componentProps": {
                "maxLength": 50
              }
            },
            {
              "component": "select",
              "key": "spaceId",
              "label": "测试空间",
              "disabled": true,
							"componentProps": {
                "options": ` + string(testSpace) +
		`}
            },
            {
              "key": "owners",
              "label": "负责人",
              "required": true,
              "component": "memberSelector",
              "componentProps": {
                "mode": "multiple",
                "scopeType": "project"
              }
            }
          ]
        }`

	var propsI interface{}
	if err := json.Unmarshal([]byte(props), &propsI); err != nil {
		logrus.Errorf("init props name=testplan component=formModal propsType=UpdateTestPlan err: errMsg: %v", err)
	}

	return propsI
}
