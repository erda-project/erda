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
