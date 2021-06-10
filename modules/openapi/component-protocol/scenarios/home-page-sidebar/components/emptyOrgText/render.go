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

package emptyOrgText

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-sidebar/i18n"
)

const (
	DefaultFontSize   = 16
	DefaultLineHeight = 24
	DefaultType       = "TextGroup"
)

func RenderCreator() protocol.CompRender {
	return &EmptyOrgText{}
}

type EmptyOrgText struct {
	ctxBdl     protocol.ContextBundle
	Type       string               `json:"type"`
	Props      props                `json:"props"`
	Operations map[string]Operation `json:"operations"`
}

type props struct {
	Visible bool          `json:"visible"`
	Align   string        `json:"align"`
	Value   []interface{} `json:"value"`
}

type Command struct {
	Key     string `json:"key"`
	Target  string `json:"target"`
	JumpOut bool   `json:"jumpOut"`
	Visible bool   `json:"visible"`
}

type Operation struct {
	Command Command `json:"command"`
	Key     string  `json:"key"`
	Reload  bool    `json:"reload"`
	Show    bool    `json:"show"`
}

func (this *EmptyOrgText) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *EmptyOrgText) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}
	i18nLocale := this.ctxBdl.Bdl.GetLocale(this.ctxBdl.Locale)
	this.Type = DefaultType
	this.Props.Align = "center"
	var visible bool
	if this.ctxBdl.Identity.OrgID == "" {
		visible = true
	}
	this.Props.Visible = visible

	this.Props.Value = make([]interface{}, 0)
	this.Props.Value = append(this.Props.Value, map[string]interface{}{
		"props": map[string]interface{}{
			"renderType": "linkText",
			"visible":    visible,
			"value": map[string]interface{}{
				"text": []interface{}{map[string]interface{}{
					"text":         i18nLocale.Get(i18n.I18nKeyOrgHowAdded),
					"operationKey": "toJoinOrgDoc",
				}},
			},
		},
	})
	this.Props.Value = append(this.Props.Value, map[string]interface{}{
		"props": map[string]interface{}{
			"renderType": "linkText",
			"visible":    visible,
			"value": map[string]interface{}{
				"text": []interface{}{map[string]interface{}{
					"text":         i18nLocale.Get(i18n.I18nKeyOrgBrosePublic),
					"operationKey": "toPublicOrgPage",
				}},
			},
		},
	})
	this.Operations = make(map[string]Operation)
	this.Operations["toJoinOrgDoc"] = Operation{
		Command: Command{
			Key:     "goto",
			Target:  "https://docs.erda.cloud/1.0/manual/platform-design.html#%E7%A7%9F%E6%88%B7-%E4%BC%81%E4%B8%9A",
			JumpOut: true,
			Visible: visible,
		},
		Key:    "click",
		Reload: false,
		Show:   false,
	}
	this.Operations["toPublicOrgPage"] = Operation{
		Command: Command{
			Key:     "goto",
			Target:  "orgList",
			JumpOut: true,
			Visible: visible,
		},
	}
	return nil
}
