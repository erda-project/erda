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

package clusteraddbutton

import (
	"context"

	"github.com/erda-project/erda/apistructs"

	edgesite "github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-site"
)

func (c *ComponentConfigsetAddButton) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	component.Props = edgesite.StructToMap(apistructs.EdgeButtonProps{
		Type: "primary",
		Text: "新建配置集",
	})
	component.Operations = apistructs.EdgeOperations{
		"click": apistructs.EdgeOperation{
			Key:    "addCluster",
			Reload: false,
			Command: apistructs.EdgeJumpCommand{
				Key: "set",
				State: apistructs.EdgeJumpCommandState{
					Visible: true,
				},
				Target: "configSetFormModal",
			},
		},
	}
	return nil
}
