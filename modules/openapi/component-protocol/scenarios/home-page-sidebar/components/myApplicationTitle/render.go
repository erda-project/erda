// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
