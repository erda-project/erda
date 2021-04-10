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

package cancelExecuteButton

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct{}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	switch event.Operation {
	case "cancelExecute":

		var req apistructs.PipelineCancelRequest
		req.PipelineID = uint64(c.State["pipelineId"].(float64))
		req.UserID = bdl.Identity.UserID
		err := bdl.Bdl.CancelPipeline(req)
		if err != nil {
			return err
		}

		c.State["reloadScenesInfo"] = true
		c.Props = map[string]interface{}{
			"text":    "取消执行",
			"visible": false,
		}
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		c.Type = "Button"
		visible := true
		if _, ok := c.State["visible"]; ok {
			visible = c.State["visible"].(bool)
		}
		if _, ok := c.State["pipelineId"]; ok && visible {
			pipelineId := uint64(c.State["pipelineId"].(float64))
			if pipelineId > 0 {
				rsp, err := bdl.Bdl.GetPipeline(pipelineId)
				if err != nil {
					return err
				}
				if !rsp.Status.IsReconcilerRunningStatus() {
					visible = false
				}
			} else {
				visible = false
			}
		}
		c.Props = map[string]interface{}{
			"text":    "取消执行",
			"visible": visible,
		}
		c.Operations = map[string]interface{}{
			"click": map[string]interface{}{
				"key":    "cancelExecute",
				"reload": true,
			},
		}
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
