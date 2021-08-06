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

package refreshButton

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct{}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	switch event.Operation {
	case apistructs.ClickOperation:
		c.State = map[string]interface{}{
			"reloadScenesInfo": true,
		}
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		c.Type = "Button"
		c.Props = map[string]interface{}{
			"text":       "刷新",
			"prefixIcon": "shuaxin",
		}
		c.Operations = map[string]interface{}{
			"click": map[string]interface{}{
				"key":    "refresh",
				"reload": true,
			},
		}
	}
	return nil
}

func pipelineShowRefresh(pipelineIDObject interface{}, bdl *bundle.Bundle) bool {

	if pipelineIDObject == nil {
		return false
	}

	pipelineID, ok := pipelineIDObject.(float64)
	if !ok {
		return false
	}

	var req apistructs.PipelineDetailRequest
	req.PipelineID = uint64(pipelineID)
	req.SimplePipelineBaseResult = true

	dto, err := bdl.GetPipelineV2(req)
	if err != nil {
		return false
	}

	if dto == nil || dto.Status.IsEndStatus() {
		return false
	}
	return true

}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
