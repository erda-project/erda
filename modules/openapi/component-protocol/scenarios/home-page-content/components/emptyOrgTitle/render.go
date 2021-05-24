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
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-content/i18n"
)

type EmptyOrgTitle struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  Props  `json:"props"`
}

type Props struct {
	Visible bool   `json:"visible"`
	Title   string `json:"title"`
	Level   int    `json:"level"`
}

func (this *EmptyOrgTitle) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (e *EmptyOrgTitle) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := e.SetCtxBundle(ctx); err != nil {
		return err
	}
	e.Type = "Title"
	if e.ctxBdl.Identity.OrgID == "" {
		e.Props.Visible = true
	}
	i18nLocale := e.ctxBdl.Bdl.GetLocale(e.ctxBdl.Locale)
	e.Props.Title = i18nLocale.Get(i18n.I18nKeyOrgEmpty)
	e.Props.Level = 2
	return nil
}

func RenderCreator() protocol.CompRender {
	return &EmptyOrgTitle{}
}
