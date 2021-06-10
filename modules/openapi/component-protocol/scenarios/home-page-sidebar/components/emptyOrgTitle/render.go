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

package emptyOrgTitle

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-sidebar/i18n"
)

type EmptyOrgTitle struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  Props  `json:"props"`
}

type Props struct {
	Visible bool          `json:"visible"`
	Align   string        `json:"align"`
	Value   []interface{} `json:"value"`
}

func (c *EmptyOrgTitle) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	c.ctxBdl = bdl
	return nil
}

func (e *EmptyOrgTitle) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := e.SetCtxBundle(ctx); err != nil {
		return err
	}

	i18nLocale := e.ctxBdl.Bdl.GetLocale(e.ctxBdl.Locale)
	e.Type = "TextGroup"
	var visible bool
	if e.ctxBdl.Identity.OrgID == "" {
		visible = true
	}
	e.Props.Visible = visible
	e.Props.Align = "center"
	e.Props.Value = []interface{}{
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "text",
				"visible":    visible,
				"value":      i18nLocale.Get(i18n.I18nKeyOrgNoAdded),
			},
		},
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &EmptyOrgTitle{}
}
