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

package emptyOrgContent

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-content/i18n"
	"github.com/erda-project/erda/modules/openapi/conf"
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
	Title         string          `json:"title,omitempty"`
	GapSize       string          `json:"gapSize,omitempty"`
	Visible       bool            `json:"visible"`
	Value         interface{}     `json:"value"`
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
	var createOrgVisible bool
	if e.ctxBdl.Identity.OrgID == "" {
		visible = true
		if conf.CreateOrgEnabled() {
			createOrgVisible = true
		}
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
		role = i18nLocale.Get(i18n.I18nKeyAdmin)
	} else {
		role = i18nLocale.Get(i18n.I18nKeyNewMember)
	}
	e.Props.Value = make([]Value, 0)
	e.Props.Value = append(e.Props.Value, Value{
		Props: PropValue{
			RenderType: "text",
			Visible:    visible,
			Value:      fmt.Sprintf("%s%s%s", i18nLocale.Get(i18n.I18nKeyOrgBelow), role, i18nLocale.Get(i18n.I18nKeyOrgQuickKnow)),
		},
		GapSize: "large",
	}, Value{
		Props: PropValue{
			RenderType: "linkText",
			Title:      fmt.Sprintf("* %s", i18nLocale.Get(i18n.I18nKeyOrgCreate)),
			Visible:    createOrgVisible,
			GapSize:    "small",
			Value: map[string]interface{}{
				"text": []interface{}{
					map[string]interface{}{
						"text": i18nLocale.Get(i18n.I18nKeyByLeft),
					},
					map[string]interface{}{
						"text":    i18nLocale.Get(i18n.I18nKeyOrgCreate),
						"withTag": true,
						"tagStyle": map[string]string{
							"backgroundColor": "#6A549E",
							"color":           "#ffffff",
							"margin":          "0 12px",
							"padding":         "4px 15px",
							"borderRadius":    "3px",
						},
					},
					map[string]interface{}{
						"text": i18nLocale.Get(i18n.I18nKeyOrgCreateQuick),
					},
				},
			},
		},
		GapSize: "large",
	},
		Value{
			Props: PropValue{
				RenderType: "text",
				Visible:    visible,
				Value:      i18nLocale.Get(i18n.I18nKeyOrgBrowsePublic),
				StyleConfig: map[string]bool{
					"bold": true,
				},
			},
			GapSize: "small",
		}, Value{
			Props: PropValue{
				RenderType: "text",
				Visible:    visible,
				Value:      i18nLocale.Get(i18n.I18nKeyOrgMsg),
			},
			GapSize: "large",
		}, Value{
			Props: PropValue{
				RenderType: "text",
				Visible:    visible,
				Value:      i18nLocale.Get(i18n.I18nKeyOrgJoin),
				StyleConfig: map[string]bool{
					"bold": true,
				},
			},
			GapSize: "small",
		}, Value{
			Props: PropValue{
				RenderType: "text",
				Visible:    visible,
				Value:      i18nLocale.Get(i18n.I18nKeyOrgJoinMsg),
			},
			GapSize: "large",
		}, Value{
			Props: PropValue{
				RenderType: "text",
				Visible:    visible,
				Value:      i18nLocale.Get(i18n.I18nKeyOrgJoinAfter),
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
