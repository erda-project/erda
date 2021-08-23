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

package refreshButton

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type Props struct {
	Visible bool   `json:"visible"`
	Text    string `json:"text"`
}

type State struct {
	AutoRefresh bool `json:"autoRefresh"`
}

type RefreshButton struct {
	Type       string                 `json:"type"`
	Props      Props                  `json:"props"`
	Operations map[string]interface{} `json:"operations"`
	State      State                  `json:"state"`
}

func (a *RefreshButton) marshal(c *apistructs.Component) error {
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

	c.State = state
	c.Type = a.Type
	c.Props = props
	return nil
}

func (r *RefreshButton) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	var autoRefresh bool
	r.Type = "Button"
	r.Props.Visible = false
	r.Operations = map[string]interface{}{
		"autoRefresh": map[string]interface{}{
			"key":         "autoRefresh",
			"reload":      true,
			"showLoading": false,
		},
	}
	switch event.Operation {
	case "autoRefresh":
		autoRefresh = true
	}
	r.State.AutoRefresh = autoRefresh
	if err := r.marshal(c); err != nil {
		return err
	}

	return nil
}

func RenderCreator() protocol.CompRender {
	return &RefreshButton{}
}
