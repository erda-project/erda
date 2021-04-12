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

package statusviewgroup

import (
	"context"
	"fmt"

	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type EdgeViewGroupState struct {
	Value string `json:"value"`
}

func (c ComponentViewGroup) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if err := c.SetBundle(bdl); err != nil {
		return err
	}

	if err := c.SetComponent(component); err != nil {
		return err
	}

	if c.component.State == nil {
		c.component.State = map[string]interface{}{}
	}

	identity := c.ctxBundle.Identity

	if event.Operation == apistructs.EdgeOperationChangeRadio || event.Operation == apistructs.InitializeOperation ||
		event.Operation == apistructs.RenderingOperation {
		err := c.OperationChangeViewGroup()
		if err != nil {
			return err
		}

		err = c.Operation(identity)
		if err != nil {
			return err
		}

		c.component.Operations = getOperations()
	}

	return nil
}

func getProps(total, success, processing, error int) apistructs.EdgeRadioProps {
	return apistructs.EdgeRadioProps{
		RadioType:   "button",
		ButtonStyle: "outline",
		Size:        "small",
		Options: []apistructs.EdgeButtonOption{
			{Text: fmt.Sprintf("全部(%d)", total), Key: "total"},
			{Text: fmt.Sprintf("成功(%d)", success), Key: "success"},
			{Text: fmt.Sprintf("部署中(%d)", processing), Key: "processing"},
			{Text: fmt.Sprintf("失败(%d)", error), Key: "error"},
		},
	}
}

func getOperations() apistructs.EdgeOperations {
	return apistructs.EdgeOperations{
		"onChange": apistructs.EdgeOperation{
			Key:    apistructs.EdgeOperationChangeRadio,
			Reload: true,
		},
	}
}
