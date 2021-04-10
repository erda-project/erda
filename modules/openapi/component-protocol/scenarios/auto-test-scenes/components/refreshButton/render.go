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

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
