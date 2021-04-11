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

package appsitebreadcrumb

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func (c ComponentBreadCrumb) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if err := c.SetBundle(bdl); err != nil {
		return err
	}

	if err := c.SetComponent(component); err != nil {
		return err
	}

	if event.Operation == apistructs.EdgeOperationClick {
	}

	c.component.Operations = getOperations()
	c.component.Data = map[string]interface{}{
		"list": []map[string]interface{}{
			{
				"key":  "appName",
				"item": "站点列表",
			},
		},
	}
	return nil
}

func getOperations() apistructs.EdgeOperations {
	return apistructs.EdgeOperations{
		apistructs.EdgeOperationClick: apistructs.EdgeOperation{
			Key:      "selectItem",
			Reload:   true,
			FillMeta: "key",
			Meta: map[string]interface{}{
				"key": "",
			},
		},
	}
}
