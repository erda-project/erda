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
