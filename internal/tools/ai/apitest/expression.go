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
	"log"
	"os"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/internal/core/openapi/legacy/component-protocol"
	"github.com/erda-project/erda/internal/core/openapi/legacy/component-protocol/pkg/autotest/step"
	"github.com/erda-project/erda/pkg/expression"
)

// 本场景入参：${{ params.scene-input-param-1 }}
// 前置接口出参：${{ outputs.52290.status }}
// 前置配置单出参：${{ outputs.52299.name }}
// 全局变量入参：${{ configs.autotest.hello }}
// mock：${{ random.integer }}
type ExpressionValues struct {
	SceneInputs            []ExpressionValue `json:"sceneInputs,omitempty"`
	PreSceneStepsOutputs   []ExpressionValue `json:"preSceneStepsOutputs,omitempty"`
	PreConfigSheetsOutputs []ExpressionValue `json:"preConfigSheetsOutputs,omitempty"`
	GlobalConfigOutputs    []ExpressionValue `json:"globalConfigOutputs,omitempty"`
	MockInputs             []ExpressionValue `json:"mockInputs,omitempty"`
}

type ExpressionValue struct {
	Name string // scene-input-param-1
	Desc string // 场景入参 1
	Expr string // ${{ params.scene-input-param-1 }}
}

func getStepExpressionValues() ExpressionValues {
	// 本场景入参
	sceneInputs, err := bdl.ListAutoTestSceneInput(apistructs.AutotestSceneRequest{
		SceneID: sceneID,
		IdentityInfo: apistructs.IdentityInfo{
			UserID: userID,
		},
	})
	if err != nil {
		log.Fatalf("failed to list scene inputs, err: %v", err)
	}
	var sceneInputValues []ExpressionValue
	for _, input := range sceneInputs {
		sceneInputValues = append(sceneInputValues, ExpressionValue{
			Name: input.Name,
			Desc: input.Description,
			Expr: expression.GenParamsRef(input.Name),
		})
	}

	// 前置接口出参
	sceneSteps, err := bdl.ListAutoTestSceneStep(apistructs.AutotestSceneRequest{
		SceneID: sceneID,
		IdentityInfo: apistructs.IdentityInfo{
			UserID: userID,
		},
	})
	if err != nil {
		log.Fatalf("failed to list scene steps, err: %v", err)
	}
	var preSceneSteps []apistructs.AutoTestSceneStep
OUTER:
	for _, serialStep := range sceneSteps {
		if serialStep.ID == sceneStepID {
			break OUTER
		}
		for _, parallelStep := range serialStep.Children {
			if parallelStep.ID == sceneStepID {
				break OUTER
			}
		}
		preSceneSteps = append(preSceneSteps, serialStep)
		preSceneSteps = append(preSceneSteps, serialStep.Children...)
	}
	preSceneStepsOutputMap, err := step.GetStepOutPut(preSceneSteps)
	if err != nil {
		log.Fatalf("failed to get pre scene steps output, err: %v", err)
	}
	var preSceneStepOutputValues []ExpressionValue
	for stepIDStr, m := range preSceneStepsOutputMap {
		// stepID: #52290-根据语言和接入方式获取接入文档
		for k, v := range m {
			exprValue := ExpressionValue{
				Name: k,
				Desc: stepIDStr + ", " + k,
				Expr: v,
			}
			preSceneStepOutputValues = append(preSceneStepOutputValues, exprValue)
		}
	}

	// 前置配置单出参
	gs := &cptype.GlobalStateData{}
	preConfigSheetsOutputMap, err := step.GetConfigSheetStepOutPut(preSceneSteps, bdl, gs)
	if err != nil {
		log.Fatalf("failed to get pre config sheet steps output, err: %v", err)
	}
	if errStr := (*gs)[protocol.GlobalInnerKeyError.String()]; errStr != nil {
		log.Fatalf("failed to get pre config sheet steps output, err: %v", errStr)
	}
	var preConfigSheetOutputValues []ExpressionValue
	for stepIDStr, m := range preConfigSheetsOutputMap {
		for k, v := range m {
			exprValue := ExpressionValue{
				Name: k,
				Desc: stepIDStr + ", " + k,
				Expr: v,
			}
			preConfigSheetOutputValues = append(preConfigSheetOutputValues, exprValue)
		}
	}

	// 全局变量入参
	globalConfigReq := apistructs.AutoTestGlobalConfigListRequest{
		Scope:   "project-autotest-testcase",
		ScopeID: os.Getenv("PROJECT_ID"),
		IdentityInfo: apistructs.IdentityInfo{
			UserID: userID,
		},
	}
	globalConfigs, err := bdl.ListAutoTestGlobalConfig(globalConfigReq)
	if err != nil {
		log.Fatalf("failed to list global configs, err: %v", err)
	}
	var globalConfigOutputValues []ExpressionValue
	for _, cfg := range globalConfigs {
		for k, v := range cfg.APIConfig.Global {
			value := ExpressionValue{
				Name: k,
				Desc: v.Desc,
				Expr: expression.GenAutotestConfigParams(k),
			}
			globalConfigOutputValues = append(globalConfigOutputValues, value)
		}
	}

	// mock
	var mockInputValues []ExpressionValue
	for _, v := range expression.MockString {
		mockInputValues = append(mockInputValues, ExpressionValue{
			Name: v,
			Desc: v,
			Expr: expression.GenRandomRef(v),
		})
	}

	return ExpressionValues{
		SceneInputs:            sceneInputValues,
		PreSceneStepsOutputs:   preSceneStepOutputValues,
		PreConfigSheetsOutputs: preConfigSheetOutputValues,
		GlobalConfigOutputs:    globalConfigOutputValues,
		MockInputs:             mockInputValues,
	}
}
