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

package sitenamefilter

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-app-site/i18n"
	i18r "github.com/erda-project/erda/pkg/i18n"
)

func (c *ComponentSiteNameFilter) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if err := c.SetBundle(bdl); err != nil {
		return err
	}
	i18nLocale := c.ctxBundle.Bdl.GetLocale(c.ctxBundle.Locale)
	if err := c.SetComponent(component); err != nil {
		return err
	}

	if c.component.State == nil {
		c.component.State = make(map[string]interface{})
	}

	if event.Operation.String() == apistructs.EdgeOperationFilter {
		err := c.OperationFilter()
		if err != nil {
			return err
		}
	}

	c.component.Operations = getOperations()
	c.component.Props = getProps()
	c.component.State["conditions"] = getStateConditions(i18nLocale)
	return nil
}

func getOperations() apistructs.EdgeOperations {
	return apistructs.EdgeOperations{
		"filter": apistructs.EdgeOperation{
			Key:    "filter",
			Reload: true,
		},
	}
}
func getStateConditions(lr *i18r.LocaleResource) []apistructs.EdgeConditions {
	return []apistructs.EdgeConditions{
		{
			Fixed:       true,
			Key:         "condition",
			Label:       lr.Get(i18n.I18nKeyName),
			Type:        "input",
			Placeholder: lr.Get(i18n.I18nKeySearchbyname),
		},
	}
}

func getProps() map[string]interface{} {
	return map[string]interface{}{
		"delay": 1000,
	}
}
