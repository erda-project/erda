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

package sitepreview

import (
	"context"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-site/i18n"
	i18r "github.com/erda-project/erda/pkg/i18n"
)

func (c *ComponentSitePreview) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if err := c.SetBundle(bdl); err != nil {
		return err
	}

	if err := c.SetComponent(component); err != nil {
		return err
	}

	identity := c.ctxBundle.Identity

	if event.Operation == apistructs.RenderingOperation {
		if err := c.OperationRendering(identity); err != nil {
			return fmt.Errorf("oepration rendering error: %v", err)
		}

	}

	return nil
}

func getProps(lr *i18r.LocaleResource) map[string]interface{} {
	return map[string]interface{}{
		"render": []PropsRender{
			{Type: "Desc", DataIndex: "siteName", Props: map[string]interface{}{"title": lr.Get(i18n.I18nKeySite)}},
			{Type: "Desc", DataIndex: "firstStep", Props: map[string]interface{}{"title": lr.Get(i18n.I18nKeyStep1)}},
			{Type: "Desc", DataIndex: "secondStep", Props: map[string]interface{}{"title": lr.Get(i18n.I18nKeyStep2)}},
			{Type: "FileEditor", DataIndex: "operationCode", Props: map[string]interface{}{"actions": map[string]interface{}{"copy": true}}},
		},
	}
}
