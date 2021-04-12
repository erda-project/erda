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

package fileConfig

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct {
	State State `json:"state"`
}

type State struct {
	ActiveKey apistructs.TestPlanActiveKey `json:"activeKey"`
	Visible   bool                         `json:"visible"`
}

func (ca *ComponentAction) GenComponentState(c *apistructs.Component) error {
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
	ca.State = state
	return nil
}

func (ca *ComponentAction) marshal(c *apistructs.Component) error {
	// state
	stateValue, err := json.Marshal(ca.State)
	if err != nil {
		return err
	}
	var stateMap map[string]interface{}
	err = json.Unmarshal(stateValue, &stateMap)
	if err != nil {
		return err
	}
	c.State = stateMap
	return nil
}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := ca.GenComponentState(c); err != nil {
		return err
	}
	props := make(map[string]interface{})
	if ca.State.ActiveKey == apistructs.ConfigTestPlanActiveKey {
		props["visible"] = true
		ca.State.Visible = true
	} else {
		props["visible"] = false
		ca.State.Visible = false
	}

	c.Props = props

	if err := ca.marshal(c); err != nil {
		return err
	}

	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
