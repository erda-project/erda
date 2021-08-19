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
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-configSet-item/i18n"
	edgesite "github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-site"
)

func (c *ComponentClusterAddButton) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if err := c.SetBundle(bdl); err != nil {
		return err
	}
	i18nLocale := c.ctxBundle.Bdl.GetLocale(c.ctxBundle.Locale)
	if component.State == nil {
		component.State = make(map[string]interface{})
	}

	if event.Operation.String() == apistructs.EdgeOperationAddCluster {
		component.State["configItemFormModalVisible"] = true
		component.State["formClear"] = true
	}

	component.Props = edgesite.StructToMap(apistructs.EdgeButtonProps{
		Type: "primary",
		Text: i18nLocale.Get(i18n.I18nKeyCreateConfigItem),
	})
	component.Operations = apistructs.EdgeOperations{
		"click": apistructs.EdgeOperation{
			Key:    apistructs.EdgeOperationAddCluster,
			Reload: true,
			Command: apistructs.EdgeJumpCommand{
				Key:    "set",
				Target: "configItemFormModal",
				State: apistructs.EdgeJumpCommandState{
					Visible: true,
				},
			},
		},
	}
	return nil
}
