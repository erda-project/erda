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

package addScenesSetButton

import (
	"context"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct{}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	switch event.Operation {
	case apistructs.ExecuteAddApiOperationKey:
		resp, err := bdl.Bdl.GetTestPlanV2(uint64(c.State["testPlanId"].(float64)))
		if err != nil {
			return err
		}
		steps := resp.Data.Steps
		var lastStepID uint64
		if len(steps) > 0 {
			lastStepID = steps[len(steps)-1].ID
		}

		var req apistructs.TestPlanV2StepAddRequest
		req.UserID = bdl.Identity.UserID
		req.TestPlanID = uint64(c.State["testPlanId"].(float64))
		req.PreID = lastStepID
		stepID, err := bdl.Bdl.CreateTestPlansV2Step(req)
		if err != nil {
			return err
		}
		c.State["showScenesSetDrawer"] = true
		fmt.Println(stepID)
		fmt.Printf("\n\n\\n\n\\n\n\\n\n\\n\n\n")

		c.State["testPlanStepId"] = stepID
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		c.Type = "Button"
		c.Props = map[string]interface{}{
			"text": "+ 场景集",
		}
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
