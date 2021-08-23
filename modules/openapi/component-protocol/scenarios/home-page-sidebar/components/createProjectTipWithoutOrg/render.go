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

package createProjectTipWithoutOrg

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-sidebar/i18n"
	"github.com/erda-project/erda/pkg/strutil"
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
				"target":  strutil.Concat("https://docs.erda.cloud/", version.Version, "/manual/platform-design.html#%E7%A7%9F%E6%88%B7-%E7%BB%84%E7%BB%87"),
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
