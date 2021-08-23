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
