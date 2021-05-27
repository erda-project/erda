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

package myApplicationTitle

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-sidebar/i18n"
)

type MyApplicationTitle struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  Props  `json:"props"`
}

type Props struct {
	Visible        bool   `json:"visible"`
	Title          string `json:"title"`
	Level          int    `json:"level"`
	NoMarginBottom bool   `json:"noMarginBottom"`
	ShowDivider    bool   `json:"showDivider"`
}

func (t *MyApplicationTitle) setProps() {
	i18nLocale := t.ctxBdl.Bdl.GetLocale(t.ctxBdl.Locale)
	t.Props.Visible = true
	t.Props.Title = i18nLocale.Get(i18n.I18nKeyAppName)
	t.Props.Level = 1
	t.Props.NoMarginBottom = true
	t.Props.ShowDivider = true
}

func (this *MyApplicationTitle) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (t *MyApplicationTitle) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := t.SetCtxBundle(ctx); err != nil {
		return err
	}
	t.Type = "Title"
	t.setProps()
	return nil
}

func RenderCreator() protocol.CompRender {
	return &MyApplicationTitle{}
}
