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

package apiEditor

import (
	"encoding/json"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/expression"
)

const mockInputStr string = `{
	"label": "mock",
	"value": "mock",
	"isLeaf": false,
	"children": [{
		"label": "string",
		"value": "${{ random.string }}",
		"isLeaf": true
	}, {
		"label": "integer",
		"value": "${{ random.integer }}",
		"isLeaf": true
	}, {
		"label": "float",
		"value": "${{ random.float }}",
		"isLeaf": true
	}, {
		"label": "boolean",
		"value": "${{ random.boolean }}",
		"isLeaf": true
	}, {
		"label": "upper",
		"value": "${{ random.upper }}",
		"isLeaf": true
	}, {
		"label": "lower",
		"value": "${{ random.lower }}",
		"isLeaf": true
	}, {
		"label": "mobile",
		"value": "${{ random.mobile }}",
		"isLeaf": true
	}, {
		"label": "digital_letters",
		"value": "${{ random.digital_letters }}",
		"isLeaf": true
	}, {
		"label": "letters",
		"value": "${{ random.letters }}",
		"isLeaf": true
	}, {
		"label": "character",
		"value": "${{ random.character }}",
		"isLeaf": true
	}, {
		"label": "timestamp",
		"value": "${{ random.timestamp }}",
		"isLeaf": true
	}, {
		"label": "timestamp_hour",
		"value": "${{ random.timestamp_hour }}",
		"isLeaf": true
	}, {
		"label": "timestamp_ns",
		"value": "${{ random.timestamp_ns }}",
		"isLeaf": true
	}, {
		"label": "timestamp_ns_hour",
		"value": "${{ random.timestamp_ns_hour }}",
		"isLeaf": true
	}, {
		"label": "date",
		"value": "${{ random.date }}",
		"isLeaf": true
	}, {
		"label": "date_day",
		"value": "${{ random.date_day }}",
		"isLeaf": true
	}, {
		"label": "datetime",
		"value": "${{ random.datetime }}",
		"isLeaf": true
	}, {
		"label": "datetime_hour",
		"value": "${{ random.datetime_hour }}",
		"isLeaf": true
	}]
}`

const props1 string = `{
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
        "width": 100,
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
        "width": 100,
        "name": "key",
        "render": {
          "required": true,
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
const prop4 string = `}`

func genProps(input, execute string) interface{} {
	var propsI interface{}
	if err := json.Unmarshal([]byte(props1+input+props2+input+props3+execute+prop4), &propsI); err != nil {
		logrus.Errorf("init props name=testplan component=formModal propsType=CreateTestPlan err: errMsg: %v", err)
	}

	return propsI
}

// Input 接口入参结构体
type Input struct {
	Label    string  `json:"label"`
	Value    string  `json:"value"`
	IsLeaf   bool    `json:"isLeaf"`
	Children []Input `json:"children"`
}

func genMockInput() Input {
	var mockInput Input
	if err := json.Unmarshal([]byte(mockInputStr), &mockInput); err != nil {
		logrus.Errorf("init mockInput name=testplan component=formModal err: errMsg: %v", err)
	}

	return mockInput
}

var mockInput = genMockInput()

// APISpec step.value 的json解析
type APISpec struct {
	APIInfo apistructs.APIInfoV2 `json:"apiSpec"`
}

func genEmptyAPISpecStr() (APISpec, string) {
	var emptySpec APISpec
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

func GetStepOutPut(steps []apistructs.AutoTestSceneStep) (map[string]map[string]string, error) {
	var value APISpec
	outputs := make(map[string]map[string]string, 0)
	for _, step := range steps {
		if step.Type == apistructs.StepTypeAPI {
			if step.Value == "" {
				step.Value = "{}"
			}
			err := json.Unmarshal([]byte(step.Value), &value)
			if err != nil {
				return nil, err
			}
			if len(value.APIInfo.OutParams) == 0 {
				continue
			}
			outputs[step.Name] = make(map[string]string, 0)
			for _, v := range value.APIInfo.OutParams {
				outputs[step.Name][v.Key] = expression.LeftPlaceholder + " outputs." + strconv.Itoa(int(step.ID)) + "." + v.Key + " " + expression.RightPlaceholder
			}
		}
	}
	return outputs, nil
}
