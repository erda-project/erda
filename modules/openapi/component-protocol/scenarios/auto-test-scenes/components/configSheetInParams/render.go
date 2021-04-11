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

package configSheetInParams

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type ComponentAction struct {
	CtxBdl     protocol.ContextBundle
	State      state                 `json:"state"`
	Props      props                 `json:"props"`
	Operations map[string]operations `json:"operations"`
	Type       string                `json:"type"`
}

type props struct {
	Fields  []interface{} `json:"fields"`
	Title   string        `json:"title"`
	Visible bool          `json:"visible"`
}
type state struct {
	FormData      interface{} `json:"formData"`
	StepID        uint64      `json:"stepId"`
	ConfigSheetID string      `json:"configSheetId"`
	Visible       bool        `json:"visible"`
}
type operations struct {
	Key      string                 `json:"key"`
	Reload   bool                   `json:"reload"`
	FillMeta string                 `json:"fillMeta"`
	Meta     interface{}            `json:"meta"`
	Command  map[string]interface{} `json:"command"`
}

func (a *ComponentAction) marshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(a.State)
	if err != nil {
		return err
	}
	var state map[string]interface{}
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}

	propValue, err := json.Marshal(a.Props)
	if err != nil {
		return err
	}
	var props map[string]interface{}
	err = json.Unmarshal(propValue, &props)
	if err != nil {
		return err
	}

	operationsValue, err := json.Marshal(a.Operations)
	if err != nil {
		return err
	}
	var operations map[string]interface{}
	err = json.Unmarshal(operationsValue, &operations)
	if err != nil {
		return err
	}

	c.State = state
	c.Type = a.Type
	c.Props = props
	c.Operations = operations
	return nil
}

func (a *ComponentAction) unmarshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	var state state
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}

	propValue, err := json.Marshal(c.Props)
	if err != nil {
		return err
	}
	var prop props
	err = json.Unmarshal(propValue, &prop)
	if err != nil {
		return err
	}

	var treeData []interface{}
	dataJson, err := json.Marshal(c.Data["treeData"])
	if err != nil {
		return err
	}
	err = json.Unmarshal(dataJson, &treeData)
	if err != nil {
		return err
	}
	a.State = state
	a.Type = c.Type
	a.Props = prop
	//a.Operations = operation
	return nil
}

func (a *ComponentAction) SetBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil {
		err := fmt.Errorf("invalie bundle")
		return err
	}
	a.CtxBdl = b
	return nil
}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component,
	scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	err := ca.unmarshal(c)
	if err != nil {
		return err
	}

	if ca.State.StepID <= 0 {
		ca.Props.Fields = nil
		ca.Props.Title = ""
		return nil
	}

	if ca.State.Visible == false {
		return nil
	}
	defer func() {
		fail := ca.marshal(c)
		if err == nil && fail != nil {
			err = fail
		}
	}()
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	err = ca.SetBundle(bdl)
	if err != nil {
		return err
	}

	switch event.Operation {
	case apistructs.OnSubmit:
		if err := ca.HandlerSubmitValue(); err != nil {
			return err
		}
	case apistructs.OnCancel:
		if err := ca.HandlerCancelValue(); err != nil {
			return err
		}
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		err := ca.handleDefault()
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *ComponentAction) handleDefault() error {
	// 选中的 step
	var configSheetID string
	var autotestGetSceneStepReq apistructs.AutotestGetSceneStepReq
	autotestGetSceneStepReq.ID = i.State.StepID
	autotestGetSceneStepReq.UserID = i.CtxBdl.Identity.UserID
	step, err := i.CtxBdl.Bdl.GetAutoTestSceneStep(autotestGetSceneStepReq)
	if err != nil {
		return err
	}
	// 选中的 step 的配置单 id
	if step.Value != "" {
		var value = make(map[string]interface{})
		err := json.Unmarshal([]byte(step.Value), &value)
		if err != nil {
			return err
		}
		i.State.FormData = value["runParams"]
		configSheetID = value["configSheetID"].(string)
	} else {
		i.State.FormData = nil
		i.Props.Fields = nil
		i.Props.Title = ""
	}

	if i.State.ConfigSheetID != "" && i.State.ConfigSheetID != "0" {
		configSheetID = i.State.ConfigSheetID
	}
	if configSheetID == "" || configSheetID == "0" {
		i.Props.Fields = nil
		i.Props.Title = ""
		i.Props.Visible = false
		i.State.Visible = false
		return nil
	} else {
		i.State.ConfigSheetID = configSheetID
	}

	// 查询配置单的入参
	var req apistructs.UnifiedFileTreeNodeGetRequest
	req.Scope = apistructs.FileTreeScopeAutoTestConfigSheet
	req.ScopeID = strconv.Itoa(int(i.CtxBdl.InParams["projectId"].(float64)))
	req.Inode = configSheetID
	req.UserID = i.CtxBdl.Identity.UserID
	orgID, err := strconv.Atoi(i.CtxBdl.Identity.OrgID)
	if err != nil {
		return err
	}
	result, err := i.CtxBdl.Bdl.GetFileTreeNode(req, uint64(orgID))
	if err != nil {
		return err
	}
	_, ok := result.Meta[apistructs.AutoTestFileTreeNodeMetaKeyPipelineYml]
	if ok {
		yml := result.Meta[apistructs.AutoTestFileTreeNodeMetaKeyPipelineYml].(string)
		pipelineYml, err := pipelineyml.New([]byte(yml))
		if err != nil {
			return err
		}
		params := pipelineYml.Spec().Params

		var fromData = map[string]interface{}{}
		for _, v := range params {

			fromData[v.Name] = v.Default

			var find = false
			for _, va := range i.Props.Fields {
				values := va.(map[string]interface{})
				if values["key"] == v.Name {
					find = true
				}
			}
			if !find {
				i.Props.Fields = append(i.Props.Fields, map[string]interface {
				}{
					"label":     v.Name,
					"component": "input",
					"required":  v.Required,
					"key":       v.Name,
				})
			}
		}

		if i.State.FormData == nil {
			i.State.FormData = fromData
		}

		if params != nil && len(params) >= 0 {
			i.Props.Title = "节点入参"
		} else {
			i.Props.Fields = nil
			i.Props.Title = ""
		}
	} else {
		i.Props.Fields = nil
		i.Props.Title = ""
	}

	i.Props.Visible = true
	i.Operations = map[string]operations{
		"submit": {
			Key:    "submit",
			Reload: true,
		},
		"cancel": {
			Key:    "cancel",
			Reload: true,
			Command: map[string]interface{}{
				"key":    "set",
				"target": "configSheetDrawer",
				"state":  map[string]interface{}{"visible": false},
			},
		},
	}
	return nil
}

func (i *ComponentAction) HandlerSubmitValue() error {
	var value = make(map[string]interface{})
	var unifiedFileTreeNodeGetRequest apistructs.UnifiedFileTreeNodeGetRequest
	unifiedFileTreeNodeGetRequest.UserID = i.CtxBdl.Identity.UserID
	unifiedFileTreeNodeGetRequest.Scope = apistructs.FileTreeScopeAutoTestConfigSheet
	unifiedFileTreeNodeGetRequest.Inode = i.State.ConfigSheetID
	orgIDInt, err := strconv.Atoi(i.CtxBdl.Identity.OrgID)
	if err != nil {
		return err
	}

	node, err := i.CtxBdl.Bdl.GetFileTreeNode(unifiedFileTreeNodeGetRequest, uint64(orgIDInt))
	if err != nil {
		return err
	}

	value["configSheetName"] = node.Name
	value["configSheetID"] = i.State.ConfigSheetID
	value["runParams"] = i.State.FormData

	valueJson, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var req apistructs.AutotestSceneRequest
	req.ID = i.State.StepID
	req.Value = string(valueJson)
	req.Name = node.Name
	req.UserID = i.CtxBdl.Identity.UserID
	_, err = i.CtxBdl.Bdl.UpdateAutoTestSceneStep(req)
	if err != nil {
		return err
	}

	i.Props.Visible = false
	i.Props.Fields = nil
	i.State.Visible = false
	i.State.StepID = 0
	i.State.ConfigSheetID = ""
	return nil
}

func (i *ComponentAction) HandlerCancelValue() error {
	i.State.Visible = false
	i.Props.Visible = false
	i.Props.Fields = nil
	i.State.StepID = 0
	i.State.ConfigSheetID = ""
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
