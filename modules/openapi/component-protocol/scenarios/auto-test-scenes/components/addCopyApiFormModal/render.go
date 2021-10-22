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

package addCopyApiFormModal

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct{}

type State struct {
	SceneID  uint64 `json:"sceneId"`
	Visible  bool   `json:"visible"`
	FormData struct {
		ApiText string `json:"apiText"`
	}
	CreateStepID uint64 `json:"createStepID"`
}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	v, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	var state State
	if err := json.Unmarshal(v, &state); err != nil {
		return err
	}
	switch event.Operation {
	case apistructs.ExecuteSubmitCopyOperationKey:
		copyText := state.FormData.ApiText
		scene, err := bdl.Bdl.GetAutoTestScene(apistructs.AutotestSceneRequest{
			IdentityInfo: apistructs.IdentityInfo{
				UserID: bdl.Identity.UserID,
			},
			AutoTestSceneParams: apistructs.AutoTestSceneParams{
				ID: state.SceneID,
			},
			Target:  -1,
			GroupID: -1,
			Type:    apistructs.StepTypeAPI,
			SceneID: state.SceneID,
		})
		if err != nil {
			return err
		}
		var req apistructs.AutotestSceneRequest
		if err := json.Unmarshal([]byte(copyText), &req); err != nil {
			return err
		}
		req.ID = 0
		req.Target = -1
		req.GroupID = -1
		req.SceneID = scene.ID
		req.SpaceID = scene.SpaceID
		req.UserID = bdl.Identity.UserID
		req.CreatorID = bdl.Identity.UserID
		req.UpdaterID = bdl.Identity.UserID
		req.APISpecID = 0
		req.RefSetID = 0
		req.PreType = ""
		req.Position = 0
		req.IsGroup = false
		stepID, err := bdl.Bdl.CreateAutoTestSceneStep(req)
		if err != nil {
			return err
		}
		c.State["createStepID"] = stepID
		c.State["visible"] = false
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		c.Props = map[string]interface{}{
			"width": 850,
			"title": "按文本添加",
			"fields": []interface{}{
				map[string]interface{}{
					"component": "textarea",
					"componentProps": map[string]interface{}{
						"autoSize": map[string]interface{}{
							"minRows": 8,
							"maxRows": 15,
						},
					},
					"key":      "apiText",
					"label":    "API文本",
					"required": true,
				},
			},
		}
		c.Operations = map[string]interface{}{
			"submit": map[string]interface{}{
				"key":    "submitCopy",
				"reload": true,
			},
		}
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
