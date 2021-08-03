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
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

const DataValueKey = "action_data_value_key"

func (a *ComponentAction) GenActionState(c *apistructs.Component) (err error) {
	sByte, err := json.Marshal(c.State)
	if err != nil {
		err = fmt.Errorf("failed to marshal action version, state:%+v, err:%v", c.State, err)
		return
	}
	state := ComponentActionState{}
	err = json.Unmarshal(sByte, &state)
	if err != nil {
		return
	}
	a.SetActionState(state)
	return
}

func GenHeaderProps(actionExt *apistructs.ExtensionVersion, versions []VersionOption) (props []apistructs.FormPropItem, err error) {
	if actionExt == nil {
		err = fmt.Errorf("empty action extension")
		return
	}
	// input
	// alias: 任务名称
	aliasInput := apistructs.FormPropItem{
		Label:     "任务名称",
		Component: "input",
		Required:  true,
		Key:       "alias",
		ComponentProps: map[string]interface{}{
			"placeholder": "请输入任务名称",
		},
		DefaultValue: actionExt.Name,
	}

	// select
	// version: 版本
	// 动态注入：所有版本选项，版本默认值
	verSelect := apistructs.FormPropItem{
		Label:     "版本",
		Component: "select",
		Required:  true,
		Key:       "version",
		ComponentProps: map[string]interface{}{
			"options": versions,
		},
		DefaultValue: actionExt.Version,
	}

	// input
	// if: 执行条件
	ifInput := apistructs.FormPropItem{
		Label:     "执行条件",
		Component: "input",
		Key:       "if",
		ComponentProps: map[string]interface{}{
			"placeholder": "请输入执行条件",
		},
	}
	props = append(props, aliasInput, verSelect, ifInput)
	return
}

func GenResourceProps(actionExt *apistructs.ExtensionVersion) (props []apistructs.FormPropItem, err error) {
	if actionExt == nil {
		err = fmt.Errorf("empty action extension")
		return
	}

	var job diceyml.Job
	diceYmlStr, err := yaml.Marshal(actionExt.Dice)
	if err != nil {
		err = fmt.Errorf("failed to marshal dice yaml, error:%v", err)
		return
	}

	diceYml, err := diceyml.New(diceYmlStr, false)
	if err != nil {
		err = fmt.Errorf("failed to parse action dice spec, error:%v", err)
		return
	}
	for k := range diceYml.Obj().Jobs {
		job = *diceYml.Obj().Jobs[k]
		break
	}

	// resourceFormGroup
	// fromGroup下会根据resource 构建一个资源列表框
	GroupResource := "resources"
	resourceFormGroup := apistructs.FormPropItem{
		Key:       GroupResource,
		Component: "formGroup",
		Group:     GroupResource,
		ComponentProps: map[string]interface{}{
			"title":         "运行资源",
			"expandable":    true,
			"defaultExpand": false,
		},
	}
	// 动态注入：cpu默认值
	resourceCpu := apistructs.FormPropItem{
		Label:          "cpu(核)",
		Component:      "inputNumber",
		Key:            GroupResource + "." + "cpu",
		Group:          GroupResource,
		ComponentProps: map[string]interface{}{},
		DefaultValue:   job.Resources.CPU,
	}
	// 动态注入：mem默认值
	resourceMem := apistructs.FormPropItem{
		Label:          "mem(MB)",
		Component:      "inputNumber",
		Key:            GroupResource + "." + "mem",
		Group:          GroupResource,
		ComponentProps: map[string]interface{}{},
		DefaultValue:   job.Resources.Mem,
	}

	props = append(props, resourceFormGroup, resourceCpu, resourceMem)
	return
}

func GenParamAndLoopProps(actionExt *apistructs.ExtensionVersion) (params []apistructs.FormPropItem, loop []apistructs.FormPropItem, err error) {
	if actionExt == nil {
		err = fmt.Errorf("empty action extension")
		return
	}
	specBytes, err := yaml.Marshal(actionExt.Spec)
	if err != nil {
		err = fmt.Errorf("failed to marshal action spec, error:%v", err)
		return
	}
	actionSpec := apistructs.ActionSpec{}
	err = yaml.Unmarshal(specBytes, &actionSpec)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal action spec, error:%v", err)
		return
	}
	params = actionSpec.FormProps
	loop = GenLoopProps(actionSpec.Loop)
	return
}

func GenLoopProps(loop *apistructs.PipelineTaskLoop) (loopFp []apistructs.FormPropItem) {
	// loopFormGroup
	// fromGroup下会根据loop 构建一个表单
	GroupLoop := "loop"
	loopFormGroup := apistructs.FormPropItem{
		Key:       GroupLoop,
		Component: "formGroup",
		Group:     GroupLoop,
		ComponentProps: map[string]interface{}{
			"title":         "循环策略",
			"expandable":    true,
			"defaultExpand": false,
		},
	}
	breakCon := apistructs.FormPropItem{
		Label:     "循环结束条件",
		Component: "input",
		Key:       GroupLoop + "." + "break",
		Group:     GroupLoop,
	}
	maxTimes := apistructs.FormPropItem{
		Label:     "最大循环次数",
		Component: "inputNumber",
		ComponentProps: map[string]interface{}{
			"precision": 0,
		},
		Key:   GroupLoop + "." + "strategy.max_times",
		Group: GroupLoop,
	}
	declineRatio := apistructs.FormPropItem{
		Label:     "衰退比例",
		Component: "inputNumber",
		Key:       GroupLoop + "." + "strategy.decline_ratio",
		Group:     GroupLoop,
		LabelTip:  "每次循环叠加间隔比例",
	}
	declineLimit := apistructs.FormPropItem{
		Label:     "衰退最大值(秒)",
		Component: "inputNumber",
		ComponentProps: map[string]interface{}{
			"precision": 0,
		},
		Key:      GroupLoop + "." + "strategy.decline_limit_sec",
		Group:    GroupLoop,
		LabelTip: "循环最大间隔时间",
	}
	interval := apistructs.FormPropItem{
		Label:     "起始间隔(秒)",
		Component: "inputNumber",
		Key:       GroupLoop + "." + "strategy.interval_sec",
		Group:     GroupLoop,
	}
	if loop != nil {
		breakCon.DefaultValue = loop.Break
		if loop.Strategy != nil {
			maxTimes.DefaultValue = loop.Strategy.MaxTimes
			declineRatio.DefaultValue = loop.Strategy.DeclineRatio
			declineLimit.DefaultValue = loop.Strategy.DeclineLimitSec
			interval.DefaultValue = loop.Strategy.IntervalSec
		}
	}

	loopFp = append(loopFp, loopFormGroup, breakCon, maxTimes, declineRatio, declineLimit, interval)
	return
}

func (a *ComponentAction) GenActionProps(name, version string) (err error) {
	actionExt, versions, err := a.QueryExtensionVersion(name, version)
	if err != nil {
		logrus.Errorf("query extension version failed, name:%s, version:%s, err:%v", name, version, err)
		return
	}
	// 默认请求时，version为空，需要以默认版本覆盖
	if version == "" {
		a.SetActionState(ComponentActionState{Version: actionExt.Version})
	}

	header, err := GenHeaderProps(actionExt, versions)
	if err != nil {
		logrus.Errorf("generate action header props failed, name:%s, version:%s, err:%v", name, version, err)
		return
	}

	params, loop, err := GenParamAndLoopProps(actionExt)
	if err != nil {
		logrus.Errorf("generate action params and loop props failed, name:%s, version:%s, err:%v", name, version, err)
		return
	}

	resource, err := GenResourceProps(actionExt)
	if err != nil {
		logrus.Errorf("generate action resource props failed, name:%s, version:%s, err:%v", name, version, err)
		return
	}

	var props []apistructs.FormPropItem
	props = append(props, header...)
	props = append(props, params...)
	props = append(props, resource...)
	props = append(props, loop...)
	a.SetActionProps(props)
	return
}

var actionTypeRender map[string]func(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, globalStateData *apistructs.GlobalStateData) (err error)
var one sync.Once

func registerActionTypeRender() {
	one.Do(func() {
		actionTypeRender = make(map[string]func(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, globalStateData *apistructs.GlobalStateData) (err error))
		actionTypeRender["manual-review"] = func(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, globalStateData *apistructs.GlobalStateData) (err error) {
			action := ctx.Value(DataValueKey).(*apistructs.PipelineYmlAction)
			params := action.Params
			if params == nil || params["processor"] == nil {
				return nil
			}
			processorJson, err := json.Marshal(params["processor"])
			if err != nil {
				return err
			}

			var processor []string
			err = json.Unmarshal(processorJson, &processor)
			if err != nil {
				return err
			}

			(*globalStateData)[protocol.GlobalInnerKeyUserIDs.String()] = processor
			return nil
		}

		actionTypeRender["testplan-run"] = func(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, globalStateData *apistructs.GlobalStateData) (err error) {
			bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
			projectId, err := strconv.Atoi(bdl.InParams["projectId"].(string))

			if err != nil {
				return err
			}
			var field []apistructs.FormPropItem
			props, ok := c.Props.(map[string]interface{})
			if !ok {
				return err
			}
			for key, val := range props {
				if key == "fields" {
					field = val.([]apistructs.FormPropItem)
					break
				}
			}

			// Add task parameters
			taskParams := apistructs.FormPropItem{
				Component: "formGroup",
				ComponentProps: map[string]interface{}{
					"title": "任务参数",
				},
				Group: "params",
				Key:   "params",
			}

			// get testplan
			testPlanRequest := apistructs.TestPlanV2PagingRequest{
				ProjectID: uint64(projectId),
			}
			testPlanRequest.UserID = bdl.Identity.UserID
			plans, err := bdl.Bdl.PagingTestPlansV2(testPlanRequest)
			if err != nil {
				return err
			}
			testPlans := make([]map[string]interface{}, 0, plans.Total)
			for _, v := range plans.List {
				testPlans = append(testPlans, map[string]interface{}{"name": fmt.Sprintf("%s-%d", v.Name, v.ID), "value": v.ID})
			}
			testPlanField := apistructs.FormPropItem{
				Label:     "测试计划",
				Component: "select",
				Required:  true,
				Key:       "params.test_plan",
				ComponentProps: map[string]interface{}{
					"options": testPlans,
				},
				Group: "params",
			}

			// get globalConfigRequest
			globalConfigRequest := apistructs.AutoTestGlobalConfigListRequest{
				ScopeID: strconv.Itoa(projectId),
				Scope:   "project-autotest-testcase",
			}
			globalConfigRequest.UserID = bdl.Identity.UserID

			globalConfigs, err := bdl.Bdl.ListAutoTestGlobalConfig(globalConfigRequest)
			if err != nil {
				return err
			}
			cms := make([]map[string]interface{}, 0, len(globalConfigs))
			for _, v := range globalConfigs {
				cms = append(cms, map[string]interface{}{"name": v.DisplayName, "value": v.Ns})
			}
			globalConfigField := apistructs.FormPropItem{
				Label:     "参数配置",
				Component: "select",
				Required:  true,
				Key:       "params.cms",
				ComponentProps: map[string]interface{}{
					"options": cms,
				},
				Group: "params",
			}

			var newField []apistructs.FormPropItem
			for _, val := range field {
				newField = append(newField, val)

				if strings.EqualFold(val.Label, "执行条件") {
					newField = append(newField, taskParams)
					newField = append(newField, testPlanField)
					newField = append(newField, globalConfigField)
				}
			}
			newProps := map[string]interface{}{
				"fields": newField,
			}
			c.Props = newProps
			return nil
		}

		actionTypeRender["testscene-run"] = func(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, globalStateData *apistructs.GlobalStateData) (err error) {
			bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
			projectId, err := strconv.ParseInt(bdl.InParams["projectId"].(string), 10, 64)
			if err != nil {
				return err
			}

			c.Operations = map[string]interface{}{
				"change": map[string]changeStruct{
					"params.test_space": {
						Key:    "changeTestSpace",
						Reload: true,
					},
					"params.test_scene_set": {
						Key:    "changeTestSet",
						Reload: true,
					},
				},
			}
			var field []apistructs.FormPropItem
			props, ok := c.Props.(map[string]interface{})
			if !ok {
				return err
			}
			for key, val := range props {
				if key == "fields" {
					field = val.([]apistructs.FormPropItem)
					break
				}
			}

			// Add task parameters
			taskParams := apistructs.FormPropItem{
				Component: "formGroup",
				ComponentProps: map[string]interface{}{
					"title": "任务参数",
				},
				Group: "params",
				Key:   "params",
			}

			spaces, err := bdl.Bdl.ListTestSpace(projectId, 500, 1)
			if err != nil {
				return err
			}
			testSpaces := make([]map[string]interface{}, 0, spaces.Total)
			for _, v := range spaces.List {
				testSpaces = append(testSpaces, map[string]interface{}{"name": v.Name, "value": v.ID})
			}
			testSceneSets := make([]map[string]interface{}, 0)

			testScenes := make([]map[string]interface{}, 0)

			spaceField := apistructs.FormPropItem{
				Label:     "测试空间",
				Component: "select",
				Required:  true,
				Key:       "params.test_space",
				ComponentProps: map[string]interface{}{
					"options": testSpaces,
				},
				Group: "params",
			}

			sceneSetField := apistructs.FormPropItem{
				Label:     "场景集",
				Component: "select",
				Required:  true,
				Key:       "params.test_scene_set",
				ComponentProps: map[string]interface{}{
					"options": testSceneSets,
				},
				Group: "params",
			}

			sceneField := apistructs.FormPropItem{
				Label:     "场景",
				Component: "select",
				Required:  true,
				Key:       "params.test_scene",
				ComponentProps: map[string]interface{}{
					"options": testScenes,
				},
				Group: "params",
			}
			formData := make(map[string]interface{})
			if strings.EqualFold(string(event.Operation), "changeTestSpace") || strings.EqualFold(string(event.Operation), "changeTestSet") {
				newMap, ok := c.State["formData"].(map[string]interface{})
				if !ok {
					return err
				}
				formData = newMap
				if strings.EqualFold(string(event.Operation), "changeTestSpace") {
					params, ok := formData["params"].(map[string]interface{})
					if !ok {
						return err
					}
					if _, ok := params["test_scene_set"]; ok {
						delete(params, "test_scene_set")
					}
					if _, ok := params["test_scene"]; ok {
						delete(params, "test_scene")
					}
					space, _ := params["test_space"].(float64)
					spaceId := uint64(space)
					sceneSetReq := apistructs.SceneSetRequest{
						SpaceID: spaceId,
					}
					sceneSetReq.UserID = bdl.Identity.UserID
					sceneSets, err := bdl.Bdl.GetSceneSets(sceneSetReq)
					if err != nil {
						return err
					}
					for _, v := range sceneSets {
						testSceneSets = append(testSceneSets, map[string]interface{}{"name": v.Name, "value": v.ID})
					}
					sceneSetField.ComponentProps = map[string]interface{}{
						"options": testSceneSets,
					}
					formData["params"] = params
				} else if strings.EqualFold(string(event.Operation), "changeTestSet") {
					params, ok := formData["params"].(map[string]interface{})
					if !ok {
						return err
					}
					if _, ok := params["test_scene"]; ok {
						delete(params, "test_scene")
					}
					space, _ := params["test_space"].(float64)
					spaceId := uint64(space)
					sceneSetReq := apistructs.SceneSetRequest{
						SpaceID: spaceId,
					}
					sceneSetReq.UserID = bdl.Identity.UserID
					sceneSets, err := bdl.Bdl.GetSceneSets(sceneSetReq)
					if err != nil {
						return err
					}
					for _, v := range sceneSets {
						testSceneSets = append(testSceneSets, map[string]interface{}{"name": v.Name, "value": v.ID})
					}
					sceneSetField.ComponentProps = map[string]interface{}{
						"options": testSceneSets,
					}

					set, _ := params["test_scene_set"].(float64)
					setId := uint64(set)
					sceneReq := apistructs.AutotestSceneRequest{
						SetID: setId,
					}
					sceneReq.UserID = bdl.Identity.UserID
					_, scenes, err := bdl.Bdl.ListAutoTestScene(sceneReq)
					if err != nil {
						return err
					}
					for _, v := range scenes {
						testScenes = append(testScenes, map[string]interface{}{"name": v.Name, "value": v.ID})
					}
					sceneField.ComponentProps = map[string]interface{}{
						"options": testScenes,
					}
					formData["params"] = params
				}
			} else {
				actionData, ok := bdl.InParams["actionData"].(map[string]interface{})
				if !ok {
					return err
				} else {
					params, ok := actionData["params"]
					if ok {
						param := params.(map[string]interface{})
						space, _ := param["test_space"].(float64)
						spaceId := uint64(space)
						sceneSetReq := apistructs.SceneSetRequest{
							SpaceID: spaceId,
						}
						sceneSetReq.UserID = bdl.Identity.UserID
						sceneSets, err := bdl.Bdl.GetSceneSets(sceneSetReq)
						if err != nil {
							return err
						}
						for _, v := range sceneSets {
							testSceneSets = append(testSceneSets, map[string]interface{}{"name": v.Name, "value": v.ID})
						}
						sceneSetField.ComponentProps = map[string]interface{}{
							"options": testSceneSets,
						}

						set, _ := param["test_scene_set"].(float64)
						setId := uint64(set)
						sceneReq := apistructs.AutotestSceneRequest{
							SetID: setId,
						}
						sceneReq.UserID = bdl.Identity.UserID
						_, scenes, err := bdl.Bdl.ListAutoTestScene(sceneReq)
						if err != nil {
							return err
						}
						for _, v := range scenes {
							testScenes = append(testScenes, map[string]interface{}{"name": v.Name, "value": v.ID})
						}
						sceneField.ComponentProps = map[string]interface{}{
							"options": testScenes,
						}
						formData["params"] = params
					}
				}
			}
			// 获取全局配置
			globalConfigRequest := apistructs.AutoTestGlobalConfigListRequest{
				ScopeID: bdl.InParams["projectId"].(string),
				Scope:   "project-autotest-testcase",
			}
			globalConfigRequest.UserID = bdl.Identity.UserID

			globalConfigs, err := bdl.Bdl.ListAutoTestGlobalConfig(globalConfigRequest)
			if err != nil {
				return err
			}
			cms := make([]map[string]interface{}, 0, len(globalConfigs))
			for _, v := range globalConfigs {
				cms = append(cms, map[string]interface{}{"name": v.DisplayName, "value": v.Ns})
			}
			globalConfigField := apistructs.FormPropItem{
				Label:     "参数配置",
				Component: "select",
				Required:  true,
				Key:       "params.cms",
				ComponentProps: map[string]interface{}{
					"options": cms,
				},
				Group: "params",
			}

			var newField []apistructs.FormPropItem
			for _, val := range field {
				newField = append(newField, val)
				if strings.EqualFold(val.Label, "执行条件") {
					newField = append(newField, taskParams)
					newField = append(newField, spaceField)
					newField = append(newField, sceneSetField)
					newField = append(newField, sceneField)
					newField = append(newField, globalConfigField)
				}
			}

			newProps := map[string]interface{}{
				"fields": newField,
			}
			c.State["formData"] = formData
			c.Props = newProps
			return nil
		}

	})
}

func (a *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, globalStateData *apistructs.GlobalStateData) (err error) {
	registerActionTypeRender()
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	err = a.SetBundle(bdl)
	if err != nil {
		return err
	}

	switch event.Operation {
	case apistructs.ChangeOperation:
		err = a.GenActionState(c)
		if err != nil {
			logrus.Errorf("generate action state failed,  err:%v", err)
			return
		}
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
	default:
		logrus.Warnf("operation [%s] not support, use default operation instead", event.Operation)
	}

	name := scenario.ScenarioKey
	version := a.GetActionVersion()
	err = a.GenActionProps(name, version)
	if err != nil {
		logrus.Errorf("generate action props failed, name:%s, version:%s, err:%v", name, version, err)
		return err
	}
	c.Props = a.Props
	cont, _ := json.Marshal(a.State)
	_ = json.Unmarshal(cont, &c.State)

	if bdl.InParams == nil {
		return nil
	}
	if bdl.InParams["actionData"] == nil {
		return nil
	}
	actionDataJson, err := json.Marshal(bdl.InParams["actionData"])
	if err != nil {
		return fmt.Errorf("failed to marshal actionData:%+v, err:%v", bdl.InParams["actionData"], err)
	}
	if len(actionDataJson) <= 0 {
		return nil
	}

	var action = &apistructs.PipelineYmlAction{}
	err = json.Unmarshal(actionDataJson, action)
	if err != nil {
		return fmt.Errorf("failed to Unmarshal actionData:%+v, err:%v", actionDataJson, err)
	}

	if action.Type == "" {
		action.Type = scenario.ScenarioKey
	}

	doFunc := actionTypeRender[action.Type]
	if doFunc == nil {
		return nil
	}

	newCtx := context.WithValue(ctx, DataValueKey, action)
	return doFunc(newCtx, c, scenario, event, globalStateData)
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
