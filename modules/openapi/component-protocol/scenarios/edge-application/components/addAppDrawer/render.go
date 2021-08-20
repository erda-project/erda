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
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-application/i18n"
)

type DrawerRendering struct {
	Visible       bool   `json:"visible"`
	OperationType string `json:"operationType"`
}

func (c *ComponentAddAppDrawer) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	var (
		drawer DrawerRendering
	)

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err := c.SetBundle(bdl); err != nil {
		return err
	}
	i18nLocale := c.ctxBundle.Bdl.GetLocale(c.ctxBundle.Locale)

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
			edgeProps.Title = i18nLocale.Get(i18n.I18nKeyEditedgeApplication)
			break
		case apistructs.EdgeOperationViewDetail:
			edgeProps.Title = i18nLocale.Get(i18n.I18nKeyEdgeApplicationDetail)
			break
		case apistructs.EdgeOperationAddApp:
			edgeProps.Title = i18nLocale.Get(i18n.I18nKeyCreateEdgeApplication)
			break
		default:
			edgeProps.Title = i18nLocale.Get(i18n.I18nKeyEdgeApplication)
		}

		component.Props = edgeProps
	}

	return nil
}
