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

package addWaitButton

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct{}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	switch event.Operation {
	case apistructs.ExecuteAddApiOperationKey:
		var req apistructs.AutotestSceneRequest
		req.Target = -1
		req.GroupID = -1
		req.Type = apistructs.StepTypeWait

		var autotestSceneRequest apistructs.AutotestSceneRequest
		autotestSceneRequest.UserID = bdl.Identity.UserID
		autotestSceneRequest.ID = uint64(c.State["sceneId"].(float64))
		autotestSceneRequest.SceneID = uint64(c.State["sceneId"].(float64))
		result, err := bdl.Bdl.GetAutoTestScene(autotestSceneRequest)
		if err != nil {
			return err
		}

		req.SceneID = result.ID
		req.SpaceID = result.SpaceID
		req.CreatorID = bdl.Identity.UserID
		req.UpdaterID = bdl.Identity.UserID
		req.UserID = bdl.Identity.UserID
		req.Value = "{\"waitTimeSec\":1}"
		stepID, err := bdl.Bdl.CreateAutoTestSceneStep(req)
		if err != nil {
			return err
		}
		c.State["createStepID"] = stepID
		c.State["showWaitEditorDrawer"] = true
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		c.Type = "Button"
		c.Props = map[string]interface{}{
			"text": "+ 等待",
		}
		c.Operations = make(map[string]interface{})
		c.Operations = map[string]interface{}{
			"click": map[string]interface{}{
				"key":    "addApi",
				"reload": true,
			},
		}
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
