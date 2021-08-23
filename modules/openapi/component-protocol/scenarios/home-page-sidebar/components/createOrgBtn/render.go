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

package createOrgBtn

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-sidebar/i18n"
	"github.com/erda-project/erda/modules/openapi/conf"
)

type CreateOrgBtn struct {
	ctxBdl     protocol.ContextBundle
	Type       string                 `json:"type"`
	Props      Props                  `json:"props"`
	Operations map[string]interface{} `json:"operations"`
}

type Props struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

func (c *CreateOrgBtn) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	c.ctxBdl = bdl
	return nil
}

func (p *CreateOrgBtn) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := p.SetCtxBundle(ctx); err != nil {
		return err
	}

	i18nLocale := p.ctxBdl.Bdl.GetLocale(p.ctxBdl.Locale)
	p.Type = "Button"
	p.Props.Text = i18nLocale.Get(i18n.I18nKeyOrgCreate)
	p.Props.Type = "primary"

	var visible bool
	if p.ctxBdl.Identity.OrgID == "" && conf.CreateOrgEnabled() {
		visible = true
	}

	p.Operations = map[string]interface{}{
		"click": map[string]interface{}{
			"key":    "addOrg",
			"reload": false,
			"command": map[string]interface{}{
				"key": "set",
				"state": map[string]bool{
					"visible": visible,
				},
				"target": "orgFormModal",
			},
		},
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &CreateOrgBtn{}
}
