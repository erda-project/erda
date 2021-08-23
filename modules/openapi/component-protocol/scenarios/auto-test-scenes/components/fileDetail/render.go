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

package fileDetail

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
	ClickType            string                          `json:"clickType"`
	ActiveKey            apistructs.ActiveKey            `json:"activeKey"`
	AutotestSceneRequest apistructs.AutotestSceneRequest `json:"autotestSceneRequest"`
	SceneId              uint64                          `json:"sceneId"`
	IsChangeScene        bool                            `json:"isChangeScene"`
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

	if ca.State.IsChangeScene {
		ca.State.ActiveKey = "fileConfig"
		ca.State.IsChangeScene = false
	}
	if event.Operation == apistructs.InitializeOperation {
		ca.State.ActiveKey = "fileConfig"
		//props["visible"] = false
	}
	if event.Operation == "changeActiveKey" {
		ca.State.ActiveKey = c.State["activeKey"].(apistructs.ActiveKey)
	}
	// props

	if ca.State.SceneId != 0 {
		props["visible"] = true
	} else {
		props["visible"] = false
	}

	props["tabMenu"] = []map[string]string{
		{"key": "fileConfig", "name": "配置信息"},
		{"key": "fileExecute", "name": "执行明细"},
	}
	c.Props = props

	c.Operations = make(map[string]interface{})
	c.Operations["onChange"] = struct {
		Key    string `json:"key"`
		Reload bool   `json:"reload"`
	}{
		Key:    "changeActiveKey",
		Reload: true,
	}
	// set state
	setState(c, ca.State)

	return json.Unmarshal([]byte(`{"onChange":{"key":"changeViewType","reload":true}}`), &c.Operations)
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}

func setState(c *apistructs.Component, state State) {
	c.State["activeKey"] = state.ActiveKey
	c.State["autotestSceneRequest"] = state.AutotestSceneRequest
	c.State["isChangeScene"] = state.IsChangeScene
	//c.State["activeKey"] = "fileExecute"
}
