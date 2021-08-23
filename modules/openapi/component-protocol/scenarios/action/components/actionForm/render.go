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

		actionTypeRender["testplan-run"] = testPlanRun

		actionTypeRender["testscene-run"] = testSceneRun
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
