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

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-site/i18n"
)

func (c *ComponentSiteAddDrawer) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	var (
		ok      bool
		visible bool
	)
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if err := c.SetBundle(bdl); err != nil {
		return err
	}
	i18nLocale := c.ctxBundle.Bdl.GetLocale(c.ctxBundle.Locale)
	if event.Operation == apistructs.RenderingOperation {
		if _, ok = component.State["visible"]; !ok {
			return nil
		}

		if visible, ok = component.State["visible"].(bool); !ok {
			return nil
		}

		if visible {
			component.Props = apistructs.EdgeDrawerProps{
				Size:  "l",
				Title: i18nLocale.Get(i18n.I18nKeyAddNode),
			}
		}
	}

	component.State = map[string]interface{}{
		"visible": visible,
	}

	return nil
}
