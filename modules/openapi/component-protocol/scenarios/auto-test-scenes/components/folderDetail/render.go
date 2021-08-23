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
