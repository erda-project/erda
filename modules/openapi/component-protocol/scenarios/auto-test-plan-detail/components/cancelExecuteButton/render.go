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
