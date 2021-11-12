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

package apiEditor

import (
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/pkg/autotest/step"
	"github.com/erda-project/erda/pkg/expression"
)

var props1 = `{
	 "loopFormField":[
		{
			"component":"formGroup",
			"key":"loop",
			"componentProps":{
				"defaultExpand": ` + LoopFormFieldDefaultExpand.string() + `,
				"expandable":true,
				"title":"循环策略"
			},
			"group":"loop"
		},
		{
			"label":"循环结束条件",
			"component":"input",
			"key":"loop.break",
			"group":"loop"
		},
		{
			"label":"最大循环次数",
			"component":"inputNumber",
			"key":"loop.strategy.max_times",
			"group":"loop"
		},
		{
			"label":"衰退比例",
			"component":"inputNumber",
			"key":"loop.strategy.decline_ratio",
			"group":"loop",
			"labelTip":"每次循环叠加间隔比例"
		},
		{
			"label":"衰退最大值(秒)",
			"component":"inputNumber",
			"key":"loop.strategy.decline_limit_sec",
			"group":"loop",
			"labelTip":"循环最大间隔时间"
		},
		{
			"label":"起始间隔(秒)",
			"component":"inputNumber",
			"key":"loop.strategy.interval_sec",
			"group":"loop"
		}
   ],
  "methodList": [
    "GET",
    "POST",
    "PUT",
    "DELETE",
    "OPTIONS",
    "PATCH",
    "COPY",
    "HEAD"
  ],
  "commonTemp": {
    "target": [
      "headers",
      "body.form"
    ],
    "temp": [
      {
        "title": "参数名",
        "key": "key",
        "width": 150,
        "render": {
          "required": true,
          "uniqueValue": true,
          "rules": [
						{
							"max": 50,
							"msg": "参数名最大长度不能超过50"
						},
            {
              "pattern": "/^[a-zA-Z0-9_-]*$/",
              "msg": "参数名为英文、数字、中划线或下划线"
            }
          ],
          "props": {
            "placeholder": "参数名"
          }
        }
      },
      {
        "title": "默认值",
        "key": "value",
        "flex": 2,
        "render": {
          "type": "inputSelect",
					"valueConvertType": "last",
          "required": true,
          "props": {
            "placeholder": "可选择表达式",
            "options": ` //[
//   {
//     "label": "前置场景",
//     "value": "$alias1.params1",
//     "isLeaf": false
//   },
//   {
//     "label": "全局参数",
//     "value": "$alias1.params2",
//     "isLeaf": true
//   }
// ]
const props2 string = `}
        }
      },
      {
        "title": "描述",
        "key": "desc",
        "width": 300,
        "render": {
          "type": "textarea",
          "required": false,
					"rules": [
						{
							"max": 1000,
							"msg": "描述最大长度不能超过1000"
						}
					],
          "props": {
            "placeholder": "描述"
          }
        }
      }
    ]
  },
  "params": {
    "temp": [
      {
        "title": "参数名",
        "key": "key",
        "width": 150,
        "name": "key",
        "render": {
          "required": true,
          "rules": [
						{
  						"max": 50,
  						"msg": "参数名最大长度不能超过50"
						},
            {
              "pattern": "/^[.a-zA-Z0-9_-]*$/",
              "msg": "参数名为英文、数字、点、中划线或下划线"
            }
          ],
          "props": {
            "placeholder": "参数名"
          }
        }
      },
      {
        "title": "默认值",
        "key": "value",
        "flex": 2,
        "name": "value",
        "render": {
          "type": "inputSelect",
					"valueConvertType": "last",
          "required": true,
          "props": {
            "placeholder": "可选择表达式",
            "options": ` //[
//   {
//     "label": "前置场景",
//     "value": "$alias1.params1",
//     "isLeaf": false,
//     "children": [
//       {
//         "label": "场景A",
//         "value": "cjA",
//         "isLeaf": false
//       },
//       {
//         "label": "场景B",
//         "value": "cjB",
//         "isLeaf": false
//       }
//     ]
//   },
//   {
//     "label": "全局参数",
//     "value": "$alias1.params2",
//     "isLeaf": true
//   }
// ]
const props3 string = `}
        }
      },
      {
        "title": "描述",
        "key": "desc",
        "placeholder": "描述",
        "name": "desc",
        "width": 300,
        "render": {
          "type": "textarea",
          "required": false,
					"rules": [
						{
							"max": 1000,
							"msg": "描述最大长度不能超过1000"
						}
					],
          "props": {
            "placeholder": "描述"
          }
        }
      }
    ],
    "showTitle": false
  },
  "headers": {
    "showTitle": false
  },
  "body": {
    "form": {
      "showTitle": false
    }
  },
  "asserts": {
    "comparisonOperators": [
      {
        "label": "大于",
        "value": ">"
      },
      {
        "label": "大于等于",
        "value": ">="
      },
      {
        "label": "等于",
        "value": "="
      },
      {
        "label": "小于等于",
        "value": "<="
      },
      {
        "label": "小于",
        "value": "<"
      },
      {
        "label": "不等于",
        "value": "!="
      },
      {
        "label": "包含",
        "value": "contains"
      },
      {
        "label": "不包含",
        "value": "not_contains"
      },
      {
        "label": "存在",
        "value": "exist"
      },
      {
        "label": "不存在",
        "value": "not_exist"
      },
      {
        "label": "为空",
        "value": "empty",
				"allowEmpty": true
      },
      {
        "label": "不为空",
        "value": "not_empty",
				"allowEmpty": true
      },
      {
        "label": "属于",
        "value": "belong"
      },
      {
        "label": "不属于",
        "value": "not_belong"
      }
    ]
  },
  "apiExecute":
`

//apiExecute: {
//  text: '执行',
//  type: 'primary',
//  disabled: true,
//  allowSave: true,
//  menu: [
//	{
//	  text: '开发环境',
//	  key: 'dev',
//	  operations: {
//		click: { key: 'execute', reload: true, meta: { env: 'dev' } },
//	  },
//	},
//	{
//	  text: '测试环境',
//	  key: 'test',
//	  operations: {
//		click: { key: 'execute', reload: true, meta: { env: 'test' } },
//	  },
//	},
//  ],
//}
//}
const props4 string = `}`

type optionKey string

// placeholder for whether the loop strategy is expanded by default
const LoopFormFieldDefaultExpand = optionKey("defaultExpand")

func (opt optionKey) string() string {
	return "{{_" + string(opt) + "_}}"
}

type replaceOption struct {
	key   optionKey
	value string
}

var defaultReplaceOptions = []replaceOption{
	{
		key:   LoopFormFieldDefaultExpand,
		value: "false",
	},
}

func genProps(input, execute string, replaceOpts ...replaceOption) interface{} {
	// because props are assembled by splicing json strings,
	// dynamic setting values can only be replaced by placeholders.
	var propsJson = props1 + input + props2 + input + props3 + execute + props4

	for _, opt := range replaceOpts {
		propsJson = strings.ReplaceAll(propsJson, opt.key.string(), opt.value)
	}

	for _, opt := range defaultReplaceOptions {
		propsJson = strings.ReplaceAll(propsJson, opt.key.string(), opt.value)
	}

	var propsI interface{}
	if err := json.Unmarshal([]byte(propsJson), &propsI); err != nil {
		logrus.Errorf("init props name=testplan component=formModal propsType=CreateTestPlan err: errMsg: %v", err)
	}

	return propsI
}

// Input 接口入参结构体
type Input struct {
	Label    string  `json:"label"`
	Value    string  `json:"value"`
	IsLeaf   bool    `json:"isLeaf"`
	ToolTip  string  `json:"tooltip"`
	Children []Input `json:"children"`
}

func genMockInput(bdl protocol.ContextBundle) Input {
	i18nLocale := bdl.Bdl.GetLocale(bdl.Locale)
	var mockInput Input
	mockInput.Label = "mock"
	mockInput.Value = "mock"
	mockInput.IsLeaf = false
	var children []Input
	for _, v := range expression.MockString {
		o := Input{
			Label:   v,
			Value:   expression.GenRandomRef(v),
			IsLeaf:  true,
			ToolTip: i18nLocale.Get("wb.content.autotest.scene."+v, v),
		}
		children = append(children, o)
	}
	mockInput.Children = children
	return mockInput
}

func genEmptyAPISpecStr() (step.APISpec, string) {
	var emptySpec step.APISpec
	emptySpec.APIInfo.Method = "GET"
	emptySpecBytes, err := json.Marshal(&emptySpec)
	if err != nil {
		logrus.Errorf("gen emptyAPISpec err: %v", err)
	}
	return emptySpec, string(emptySpecBytes)
}

const (
	EmptySpecID string = "undefined"
)

var (
	emptySpec, emptySpecStr = genEmptyAPISpecStr()
)
