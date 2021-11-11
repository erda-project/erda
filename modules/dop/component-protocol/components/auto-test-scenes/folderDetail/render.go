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

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentAction struct {
	base.DefaultProvider
	State State `json:"state"`
}

type State struct {
	SetId   uint64 `json:"setId"`
	SceneId uint64 `json:"sceneId"`
}

func init() {
	base.InitProviderWithCreator("auto-test-scenes", "folderDetail",
		func() servicehub.Provider { return &ComponentAction{} })
}

func (ca *ComponentAction) RenderState(c *cptype.Component) error {
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

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
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

func setState(c *cptype.Component, state State) {
	c.State["setId"] = state.SetId
	c.State["sceneId"] = state.SceneId
}
