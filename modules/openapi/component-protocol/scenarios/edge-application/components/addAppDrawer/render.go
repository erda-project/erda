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

package siteadddrawer

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
)

type DrawerRendering struct {
	Visible       bool   `json:"visible"`
	OperationType string `json:"operationType"`
}

func (c *ComponentAddAppDrawer) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	var (
		drawer DrawerRendering
	)

	if component.State == nil {
		component.State = map[string]interface{}{}
	}

	if event.Operation == apistructs.RenderingOperation {

		jsonData, err := json.Marshal(component.State)
		if err != nil {
			return err
		}

		if err = json.Unmarshal(jsonData, &drawer); err != nil {
			return err
		}

		component.State = map[string]interface{}{
			"visible": drawer.Visible,
		}

		edgeProps := apistructs.EdgeDrawerProps{
			Size: "l",
		}

		switch drawer.OperationType {
		case apistructs.EdgeOperationUpdate:
			edgeProps.Title = "编辑边缘应用"
			break
		case apistructs.EdgeOperationViewDetail:
			edgeProps.Title = "边缘应用详情"
			break
		case apistructs.EdgeOperationAddApp:
			edgeProps.Title = "发布边缘应用"
			break
		default:
			edgeProps.Title = "边缘应用"
		}

		component.Props = edgeProps
	}

	return nil
}
