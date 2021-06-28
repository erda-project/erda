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

package createProjectTipWithoutOrg

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-sidebar/i18n"
)

type CreateProjectTipWithoutOrg struct {
	ctxBdl     protocol.ContextBundle
	Type       string                 `json:"type"`
	Props      Props                  `json:"props"`
	Operations map[string]interface{} `json:"operations"`
}

type Props struct {
	Visible    bool   `json:"visible"`
	RenderType string `json:"renderType"`
	Value      map[string]interface{}
}

func (this *CreateProjectTipWithoutOrg) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (p *CreateProjectTipWithoutOrg) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	if err := p.SetCtxBundle(ctx); err != nil {
		return err
	}

	var visible bool
	i18nLocale := p.ctxBdl.Bdl.GetLocale(p.ctxBdl.Locale)
	if p.ctxBdl.Identity.OrgID == "" {
		visible = true
	}
	p.Type = "Text"
	p.Props.RenderType = "linkText"
	p.Props.Visible = visible
	p.Props.Value = map[string]interface{}{
		"text": []interface{}{
			i18nLocale.Get(i18n.I18nKeyOrgAddFirst),
			map[string]interface{}{
				"text":         i18nLocale.Get(i18n.I18nKeyMoreContent),
				"operationKey": "toJoinOrgDoc",
			},
		},
	}
	p.Operations = map[string]interface{}{
		"toJoinOrgDoc": map[string]interface{}{
			"command": map[string]interface{}{
				"key":     "gogo",
				"target":  "https://docs.erda.cloud/1.0/manual/platform-design.html#%E7%A7%9F%E6%88%B7-%E7%BB%84%E7%BB%87",
				"jumpOut": true,
				"visible": visible,
			},
		},
		"key":    "click",
		"reload": false,
		"show":   false,
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &CreateProjectTipWithoutOrg{}
}
