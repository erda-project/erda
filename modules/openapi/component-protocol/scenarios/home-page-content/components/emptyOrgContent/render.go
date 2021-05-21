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

package emptyOrgContent

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	i18n2 "github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-content/i18n"
)

type EmptyOrgContent struct {
	ctxBdl     protocol.ContextBundle
	Type       string                 `json:"type"`
	Props      Props                  `json:"props"`
	Operations map[string]interface{} `json:"operations"`
}

type Props struct {
	Visible bool    `json:"visible"`
	Value   []Value `json:"value"`
}

type Value struct {
	Props   PropValue `json:"props"`
	GapSize string    `json:"gapSize"`
}

type PropValue struct {
	RenderType    string          `json:"renderType"`
	Visible       bool            `json:"visible"`
	Value         string          `json:"value"`
	StyleConfig   map[string]bool `json:"styleConfig,omitempty"`
	TextStyleName map[string]bool `json:"textStyleName,omitempty"`
}

func (this *EmptyOrgContent) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (e *EmptyOrgContent) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := e.SetCtxBundle(ctx); err != nil {
		return err
	}
	e.Type = "TextGroup"
	var visible bool
	if e.ctxBdl.Identity.OrgID == "" {
		visible = true
	}
	e.Props.Visible = visible
	var access bool
	var role string
	if e.ctxBdl.Identity.OrgID != "" {
		orgIntId, err := strconv.Atoi(e.ctxBdl.Identity.OrgID)
		req := &apistructs.PermissionCheckRequest{
			UserID:   e.ctxBdl.Identity.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(orgIntId),
			Resource: apistructs.ProjectResource,
			Action:   apistructs.CreateAction,
		}
		permissionRes, err := e.ctxBdl.Bdl.CheckPermission(req)
		if err != nil {
			return err
		}
		if permissionRes == nil {
			return fmt.Errorf("can not check permission for org")
		}
		access = permissionRes.Access
	}
	i18nLocale := e.ctxBdl.Bdl.GetLocale(e.ctxBdl.Locale)
	if access {
		role = i18nLocale.Get(i18n2.I18nKeyAdmin)
	} else {
		role = i18nLocale.Get(i18n2.I18nKeyNewMember)
	}
	e.Props.Value = make([]Value, 0)
	e.Props.Value = append(e.Props.Value, Value{
		Props: PropValue{
			RenderType: "text",
			Visible:    visible,
			Value:      fmt.Sprintf("%s%s%s", i18nLocale.Get(i18n2.I18nKeyOrgBelow), role, i18nLocale.Get(i18n2.I18nKeyOrgQuickKnow)),
		},
		GapSize: "large",
	}, Value{
		Props: PropValue{
			RenderType: "text",
			Visible:    visible,
			Value:      i18nLocale.Get(i18n2.I18nKeyOrgBrowsePublic),
			StyleConfig: map[string]bool{
				"bold": true,
			},
		},
		GapSize: "small",
	}, Value{
		Props: PropValue{
			RenderType: "text",
			Visible:    visible,
			Value:      i18nLocale.Get(i18n2.I18nKeyOrgMsg),
		},
		GapSize: "large",
	}, Value{
		Props: PropValue{
			RenderType: "text",
			Visible:    visible,
			Value:      i18nLocale.Get(i18n2.I18nKeyOrgJoin),
			StyleConfig: map[string]bool{
				"bold": true,
			},
		},
		GapSize: "small",
	}, Value{
		Props: PropValue{
			RenderType: "text",
			Visible:    visible,
			Value:      i18nLocale.Get(i18n2.I18nKeyOrgJoinMsg),
		},
		GapSize: "large",
	}, Value{
		Props: PropValue{
			RenderType: "text",
			Visible:    visible,
			Value:      i18nLocale.Get(i18n2.I18nKeyOrgJoinAfter),
			TextStyleName: map[string]bool{
				"fz12":            true,
				"color-text-desc": true,
			},
		},
		GapSize: "normal",
	})
	e.Operations = map[string]interface{}{
		"toSpecificProject": map[string]interface{}{
			"command": map[string]interface{}{
				"key":     "goto",
				"target":  "projectAllIssue",
				"jumpOut": true,
				"state": map[string]interface{}{
					"query": map[string]interface{}{
						"issueViewGroup__urlQuery": "eyJ2YWx1ZSI6ImthbmJhbiIsImNoaWxkcmVuVmFsdWUiOnsia2FuYmFuIjoiZGVhZGxpbmUifX0=",
					},
					"params": map[string]interface{}{
						"projectId": "",
					},
				},
				"visible": visible,
			},
			"key":    "click",
			"reload": false,
			"show":   false,
		},
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &EmptyOrgContent{}
}
