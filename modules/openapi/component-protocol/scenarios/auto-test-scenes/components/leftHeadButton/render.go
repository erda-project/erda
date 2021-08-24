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

package leftHeadButton

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-scenes/components/fileTree"
)

// SetCtxBundle 设置bundle
func (i *ComponentleftHeadButtonModal) SetCtxBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil || b.I18nPrinter == nil {
		err := fmt.Errorf("invalie context bundle")
		return err
	}
	logrus.Infof("inParams:%+v, identity:%+v", b.InParams, b.Identity)
	i.CtxBdl = b
	return nil
}

func (i *ComponentleftHeadButtonModal) marshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(i.State)
	if err != nil {
		return err
	}
	var state map[string]interface{}
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}

	// var data apistructs.ComponentData = map[string]interface{}{}
	// data["treeData"] = i.Data
	// c.Data = data
	c.State = state
	// c.Type = i.Type
	return nil
}

func (a *ComponentleftHeadButtonModal) unmarshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	var state State
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	a.State = state
	// a.Type = c.Type
	// a.Data = data
	return nil
}

func (i *ComponentleftHeadButtonModal) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	if event.Operation != apistructs.InitializeOperation && event.Operation != apistructs.RenderingOperation {
		err = i.unmarshal(c)
		if err != nil {
			return err
		}
	}

	defer func() {
		fail := i.marshal(c)
		if err == nil && fail != nil {
			err = fail
		}
	}()

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	err = i.SetCtxBundle(bdl)
	if err != nil {
		return err
	}

	err = i.GenComponentState(c)
	if err != nil {
		return err
	}

	if i.CtxBdl.InParams == nil {
		return fmt.Errorf("params is empty")
	}

	inParamsBytes, err := json.Marshal(i.CtxBdl.InParams)
	if err != nil {
		return fmt.Errorf("failed to marshal inParams, inParams:%+v, err:%v", i.CtxBdl.InParams, err)
	}

	var inParams fileTree.InParams
	if err := json.Unmarshal(inParamsBytes, &inParams); err != nil {
		return err
	}

	i.Initial()
	// i.Props =

	switch event.Operation {
	// case apistructs.InitializeOperation, apistructs.RenderingOperation:
	// 	if err := i.Initial(bdl, inParams); err != nil {
	// 		return err
	// 	}
	case apistructs.ClickAddSceneSeButtonOperationKey:
		if err := i.RenderClickButton(); err != nil {
			return err
		}
	}
	i.RenderProtocol(c, gs)
	return nil
}

func (i *ComponentleftHeadButtonModal) RenderProtocol(c *apistructs.Component, g *apistructs.GlobalStateData) {
	if c.Operations == nil {
		d := make(apistructs.ComponentOps)
		c.Operations = d
	}
	if c.State == nil {
		d := make(apistructs.ComponentData)
		c.State = d
	}
	// c.State =
	c.Operations = i.Operations
	// c.Props = i.Props
}

// GenComponentState 获取state
func (i *ComponentleftHeadButtonModal) GenComponentState(c *apistructs.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	i.State = state
	return nil
}

func (i *ComponentleftHeadButtonModal) RenderClickButton() error {
	i.State.ActionType = "ClickAddSceneSetButton"
	i.State.FormVisible = true
	return nil
}

func (i *ComponentleftHeadButtonModal) Initial() {
	var click = Operation{
		Key:    "ClickAddSceneSet",
		Reload: true,
	}
	i.Operations = map[string]interface{}{}
	i.Operations["click"] = click
}

func RenderCreator() protocol.CompRender {
	return &ComponentleftHeadButtonModal{}
}
