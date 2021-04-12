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

package executeHistoryButton

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct{}

func getProps(visible bool) map[string]interface{} {
	return map[string]interface{}{
		"text":    "执行历史",
		"visible": visible,
	}
}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	//return json.Unmarshal([]byte(`{"text":"执行历史"}`), &c.Props)
	visible := true
	if _, ok := c.State["visible"]; ok {
		visible = c.State["visible"].(bool)
	}
	c.Props = getProps(visible)
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
