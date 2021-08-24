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
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pingcap/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/pkg/expression"
)

type ApiEditor struct {
	ctxBdl protocol.ContextBundle

	State State `json:"state"`
}

func RenderCreator() protocol.CompRender {
	return &ApiEditor{}
}

// SetCtxBundle 设置bundle
func (ae *ApiEditor) SetCtxBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil || b.I18nPrinter == nil {
		err := fmt.Errorf("invalie context bundle")
		return err
	}
	logrus.Infof("inParams:%+v, identity:%+v", b.InParams, b.Identity)
	ae.ctxBdl = b
	return nil
}

func GetOpsInfo(opsData interface{}) (*Meta, error) {
	if opsData == nil {
		err := fmt.Errorf("empty operation data")
		return nil, err
	}
	var op OpMetaInfo
	cont, err := json.Marshal(opsData)
	if err != nil {
		logrus.Errorf("marshal inParams failed, content:%v, err:%v", opsData, err)
		return nil, err
	}
	err = json.Unmarshal(cont, &op)
	if err != nil {
		logrus.Errorf("unmarshal move out request failed, content:%v, err:%v", cont, err)
		return nil, err
	}
	meta := op.Meta
	return &meta, nil
}

func (ae *ApiEditor) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	tmpStepID, ok := c.State["stepId"]
	if !ok {
		return errors.New("stepId is empty")
	}
	stepID, err := strconv.ParseUint(fmt.Sprintf("%v", tmpStepID), 10, 64)
	if err != nil {
		return err
	}

	if stepID == 0 {
		ae.State.AttemptTest = AttemptTestAll{}
		return nil
	}

	// state
	ae.State.StepId = stepID

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err := ae.SetCtxBundle(bdl); err != nil {
		return err
	}
	userID := bdl.Identity.UserID
	orgIDStr := bdl.Identity.OrgID
	if orgIDStr == "" {
		return errors.New("orgID is empty")
	}
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		return err
	}
	// var orgID uint64 = 1
	sceneIDStr := fmt.Sprintf("%v", bdl.InParams["sceneId__urlQuery"])
	sceneID, err := strconv.ParseUint(sceneIDStr, 10, 64)
	if err != nil {
		return err
	}
	ae.State.SceneId = sceneID

	projecrIDStr := fmt.Sprintf("%v", bdl.InParams["projectId"])

	// 获取小试一把信息
	if _, ok := c.State["isFirstIn"]; ok {
		ae.State.IsFirstIn = c.State["isFirstIn"].(bool)
	}
	if _, ok := c.State["attemptTest"]; ok {
		mp := c.State["attemptTest"].(map[string]interface{})
		bt, err := json.Marshal(mp)
		if err != nil {
			return err
		}
		err = json.Unmarshal(bt, &ae.State.AttemptTest)
		if err != nil {
			return err
		}
	}
	// 如果是初次进入 清空小试一把信息
	if ae.State.IsFirstIn == true {
		c.State["isFirstIn"] = false
		c.State["attemptTest"] = AttemptTestAll{}
	}
	// 塞props
	// 本场景入参
	var inputs []Input
	sceneInputReq := apistructs.AutotestSceneRequest{}
	sceneInputReq.SceneID = sceneID
	sceneInputReq.UserID = userID
	sceneInputs, err := bdl.Bdl.ListAutoTestSceneInput(sceneInputReq)
	if err != nil {
		return err
	}
	children := make([]Input, 0, 0)
	for _, sip := range sceneInputs {
		children = append(children, Input{Label: sip.Name, Value: "${{ params." + sip.Name + " }}", IsLeaf: true})
	}
	inputs = append(inputs, Input{Label: "本场景入参", Value: "本场景入参", IsLeaf: false, Children: children})

	// 前置接口入参
	sceneSteps, err := bdl.Bdl.ListAutoTestSceneStep(sceneInputReq)
	if err != nil {
		return err
	}
	var steps []apistructs.AutoTestSceneStep
LABEL:
	for _, sStep := range sceneSteps {
		if sStep.ID == stepID {
			break
		}
		for _, pStep := range sStep.Children {
			if pStep.ID == stepID {
				break LABEL
			}
		}
		steps = append(steps, sStep)
		steps = append(steps, sStep.Children...)
	}
	maps, err := GetStepOutPut(steps)
	if err != nil {
		return err
	}
	stepChildren1 := make([]Input, 0, 0)
	for k, v := range maps {
		stepChildren2 := make([]Input, 0, 0)
		for k1, v1 := range v {
			stepChildren2 = append(stepChildren2, Input{Label: k1, Value: v1, IsLeaf: true})
		}
		stepChildren1 = append(stepChildren1, Input{Label: k, Value: k, IsLeaf: false, Children: stepChildren2})
	}
	inputs = append(inputs, Input{Label: "前置接口出参", Value: "前置接口出参", IsLeaf: false, Children: stepChildren1})

	// 全局变量入参
	cfgReq := apistructs.AutoTestGlobalConfigListRequest{Scope: "project-autotest-testcase", ScopeID: projecrIDStr}
	cfgReq.UserID = userID
	cfgs, err := bdl.Bdl.ListAutoTestGlobalConfig(cfgReq)
	if err != nil {
		return err
	}
	cfgChildren0 := make([]Input, 0, 0)
	for _, cfg := range cfgs {
		// Header 是自动带上去的
		// cfgChildren1, cfgChildren2, cfgChildren3 := make([]Input, 0, 0), make([]Input, 0, 0), make([]Input, 0, 0)
		cfgChildren1, cfgChildren3 := make([]Input, 0, 0), make([]Input, 0, 0)
		// for k := range cfg.APIConfig.Header {
		// 	cfgChildren2 = append(cfgChildren2, Input{Label: k, Value: "{{" + k + "}}", IsLeaf: true})
		// }
		for _, v := range cfg.APIConfig.Global {
			cfgChildren3 = append(cfgChildren3, Input{Label: v.Name, Value: expression.GenAutotestConfigParams(v.Name), IsLeaf: true})
		}
		// cfgChildren1 = append(cfgChildren1, Input{Label: "Header", Value: "Header", IsLeaf: false, Children: cfgChildren2})
		cfgChildren1 = append(cfgChildren1, Input{Label: "Global", Value: "Global", IsLeaf: false, Children: cfgChildren3})
		cfgChildren0 = append(cfgChildren0, Input{Label: cfg.DisplayName, Value: cfg.DisplayName, IsLeaf: false, Children: cfgChildren1})
	}
	inputs = append(inputs, Input{Label: "全局变量入参", Value: "全局变量入参", IsLeaf: false, Children: cfgChildren0})

	// mock 入参
	inputs = append(inputs, genMockInput(bdl))

	inputBytes, err := json.Marshal(inputs)
	if err != nil {
		return err
	}
	executeString, err := ae.GenExecuteButton()
	if err != nil {
		return err
	}
	c.Props = genProps(string(inputBytes), executeString)
	c.State["apiEditorDrawerVisible"] = true

	switch event.Operation.String() {
	case "onChange":
		// 失焦更新
		err := ae.Save(c)
		if err != nil {
			return err
		}
		c.State["apiEditorDrawerVisible"] = false
		return nil
	case "execute":
		err := ae.Save(c)
		if err != nil {
			return err
		}
		err = ae.ExecuteApi(event.OperationData)
		if err != nil {
			return err
		}
		c.State["attemptTest"] = ae.State.AttemptTest
		return nil
	}

	// marketProto 的changeAPISpec 事件
	tmpSpecID, ok := c.State["changeApiSpecId"]
	if ok && tmpSpecID != nil && tmpSpecID != 0 {
		if _, ok := tmpSpecID.(string); ok && tmpSpecID == EmptySpecID {
			// 切换成自定义api
			updateReq := apistructs.AutotestSceneRequest{
				SceneID:   sceneID,
				Value:     emptySpecStr,
				APISpecID: 0,
				Name:      "",
			}
			updateReq.UserID = userID
			updateReq.ID = stepID
			if _, err := bdl.Bdl.UpdateAutoTestSceneStep(updateReq); err != nil {
				return err
			}

			data := map[string]interface{}{"apiSpec": emptySpec.APIInfo, "apiSpecId": nil}
			c.State["data"] = data
			c.State["changeApiSpecId"] = nil
			return nil
		}
		// 切换成另一个api集市的api
		apiSpecID, err := strconv.ParseUint(fmt.Sprintf("%v", tmpSpecID), 10, 64)
		if err != nil {
			return err
		}
		apiSpec, err := bdl.Bdl.GetAPIOperation(orgID, userID, apiSpecID)
		if err != nil {
			return err
		}

		var headers []apistructs.APIHeader
		for _, v := range apiSpec.Headers {
			headers = append(headers, apistructs.APIHeader{Key: v.Name, Value: "", Desc: v.Description})
		}
		var params []apistructs.APIParam
		for _, v := range apiSpec.Parameters {
			params = append(params, apistructs.APIParam{Key: v.Name, Value: "", Desc: v.Description})
		}
		var body apistructs.APIBody
		for _, v := range apiSpec.RequestBody {
			if v.MediaType == "application/json" {
				body.Type = apistructs.APIBodyTypeApplicationJSON2
				body.Content = v.Body.Example
				break
			}
		}
		if len(apiSpec.RequestBody) != 0 && body.Type == "" {
			body.Type = apistructs.APIBodyType(apiSpec.RequestBody[0].MediaType)
			body.Content = apiSpec.RequestBody[0].Body.Example
		}

		apiInfo := APISpec{
			APIInfo: apistructs.APIInfoV2{
				Name:      apiSpec.OperationID,
				URL:       apiSpec.Path,
				Method:    apiSpec.Method,
				Headers:   headers,
				Params:    params,
				Body:      body,
				OutParams: []apistructs.APIOutParam{},
				Asserts:   []apistructs.APIAssert{},
			},
		}

		apiInfoBytes, err := json.Marshal(&apiInfo)
		if err != nil {
			return err
		}
		// 更新步骤的value
		updateReq := apistructs.AutotestSceneRequest{
			SceneID:   sceneID,
			Value:     string(apiInfoBytes),
			APISpecID: apiSpecID,
			Name:      apiInfo.APIInfo.Name,
		}
		updateReq.UserID = userID
		updateReq.ID = stepID
		if _, err := bdl.Bdl.UpdateAutoTestSceneStep(updateReq); err != nil {
			return err
		}

		data := map[string]interface{}{"apiSpec": apiInfo.APIInfo, "apiSpecId": apiSpecID}
		c.State["data"] = data
		c.State["changeApiSpecId"] = nil

		return nil
	}

	// 第一次进来，获取step渲染
	step, err := bdl.Bdl.GetAutoTestSceneStep(apistructs.AutotestGetSceneStepReq{
		ID:     stepID,
		UserID: userID,
	})
	if err != nil {
		return err
	}
	if step.Value == "" {
		step.Value = emptySpecStr
	}
	var apiInfo APISpec
	// TODO 名字用step表里的name
	if err := json.Unmarshal([]byte(step.Value), &apiInfo); err != nil {
		return err
	}
	apiInfo.APIInfo.Name = step.Name
	c.State["data"] = map[string]interface{}{"apiSpec": apiInfo.APIInfo, "apiSpecId": step.APISpecID, "loop": apiInfo.Loop}
	return nil
}
