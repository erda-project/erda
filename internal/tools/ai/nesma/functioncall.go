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
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

const (
	fieldSubsystem         = "子系统"
	fieldFirstLevelModule  = "一级模块"
	fieldSecondLevelModule = "二级模块"
	fieldFPCountItemName   = "功能点计数项名称"
	fieldFPType            = "类别"
)

type FunctionPoint struct {
	Subsystem         string `json:"子系统"`
	FirstLevelModule  string `json:"一级模块"`
	SecondLevelModule string `json:"二级模块"`
	FPCountItemName   string `json:"功能点计数项名称"`
	FPType            string `json:"类别"`
	ReuseLevel        string `json:"复用度系数"`
	AdjustLevel       string `json:"修改类型调整系数"`
	RepeatCount       string `json:"重复计数"`
}

func (fp *FunctionPoint) ToTableLine() string {
	// split by '	'
	// 供应商门户	供应商工作台	合同	查看合同列表	EQ	低	新增	否
	ss := []string{fp.Subsystem, fp.FirstLevelModule, fp.SecondLevelModule, fp.FPCountItemName, fp.FPType, fp.ReuseLevel, fp.AdjustLevel, fp.RepeatCount}
	return strings.Join(ss, "\t")
}

var calculateFunctionPoint = openai.FunctionDefinition{
	Name:        "calculate-function-point",
	Description: "calculate function point by Nesma",
	Parameters: &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"list": {
				Type:        jsonschema.Array,
				Description: "list of function point",
				Items: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						fieldSubsystem: {
							Type:        jsonschema.String,
							Description: "the name of subs system",
						},
						fieldFirstLevelModule: {
							Type:        jsonschema.String,
							Description: "the name of first level module",
						},
						fieldSecondLevelModule: {
							Type:        jsonschema.String,
							Description: "the name of second level module",
						},
						fieldFPCountItemName: {
							Type:        jsonschema.String,
							Description: "the name of function point count item",
						},
						fieldFPType: {
							Type:        jsonschema.String,
							Description: "the type of function point",
							Enum:        []string{"ILF", "EIF", "EI", "EO", "EQ"},
						},
					},
					Required: []string{fieldSubsystem, fieldFirstLevelModule, fieldSecondLevelModule, fieldFPCountItemName, fieldFPType},
				},
			},
		},
	},
}
