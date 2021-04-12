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

package addappbutton

import (
	"context"

	"github.com/erda-project/erda/apistructs"
)

func (c ComponentAddAppButton) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if component.State == nil {
		component.State = make(map[string]interface{})
	}

	if event.Operation.String() == apistructs.EdgeOperationAddApp {
		component.State["appConfigFormVisible"] = true
		component.State["addAppDrawerVisible"] = true
		component.State["keyValueListTitleVisible"] = false
		component.State["keyValueListVisible"] = false
		component.State["formClear"] = true
		component.State["operationType"] = event.Operation
	}

	component.Props = getProps()
	component.Operations = getOperations()

	return nil
}

func getProps() apistructs.EdgeButtonProps {
	return apistructs.EdgeButtonProps{
		Type: "primary",
		Text: "发布应用",
	}
}

func getOperations() apistructs.EdgeOperations {
	return apistructs.EdgeOperations{
		"click": apistructs.EdgeOperation{
			Key:    apistructs.EdgeOperationAddApp,
			Reload: true,
			Command: apistructs.EdgeJumpCommand{
				Key: "set",
				State: apistructs.EdgeJumpCommandState{
					Visible:  true,
					ReadOnly: false,
				},
				Target: "appFormModal",
			},
		},
	}
}
