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

package folderDetail

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct {
	State State `json:"state"`
}

type State struct {
	SetId   uint64 `json:"setId"`
	SceneId uint64 `json:"sceneId"`
}

func (ca *ComponentAction) RenderState(c *apistructs.Component) error {
	var state State
	b, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &state); err != nil {
		return err
	}
	ca.State = state
	return nil
}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if c.State == nil {
		c.State = map[string]interface{}{}
	}

	if err := ca.RenderState(c); err != nil {
		return err
	}

	props := make(map[string]interface{})
	if ca.State.SceneId == 0 && ca.State.SetId != 0 {
		props["visible"] = true
	} else {
		props["visible"] = false
	}
	c.Props = props
	// set state
	setState(c, ca.State)

	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}

func setState(c *apistructs.Component, state State) {
	c.State["setId"] = state.SetId
	c.State["sceneId"] = state.SceneId
}
