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

package emptyProjectContent

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-content/i18n"
)

type EmptyProjectContent struct {
	ctxBdl protocol.ContextBundle
	Type   string                 `json:"type"`
	Props  map[string]interface{} `json:"props"`
	State  State                  `json:"state"`
}

type State struct {
	ProsNum int `json:"prosNum"`
}

func (this *EmptyProjectContent) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *EmptyProjectContent) GenComponentState(c *apistructs.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	this.State = state
	return nil
}

func (e *EmptyProjectContent) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := e.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := e.GenComponentState(c); err != nil {
		return err
	}
	e.Type = "TextGroup"
	e.Props = make(map[string]interface{})
	if e.ctxBdl.Identity.OrgID == "" {
		e.Props["visible"] = false
		return nil
	}
	var visible bool
	var access bool
	var role string
	var createProStr string
	var createProDetail interface{}
	var createProType string
	i18nLocale := e.ctxBdl.Bdl.GetLocale(e.ctxBdl.Locale)
	if e.State.ProsNum == 0 {
		visible = true
	}
	orgIDInt, err := strconv.Atoi(e.ctxBdl.Identity.OrgID)
	if err != nil {
		return err
	}
	members, err := e.ctxBdl.Bdl.ListMembers(apistructs.MemberListRequest{
		ScopeType: apistructs.OrgScope,
		ScopeID:   int64(orgIDInt),
		PageNo:    1,
		PageSize:  1000,
	})
	if err != nil {
		return fmt.Errorf("check permission failed: %v", err)
	}
	var joined bool
	for _, member := range members {
		if member.UserID == e.ctxBdl.Identity.UserID {
			joined = true
			break
		}
	}
	if !joined {
		e.Props["value"] = []interface{}{
			map[string]interface{}{
				"props": map[string]interface{}{
					"renderType": "Text",
					"visible":    visible,
					"value":      i18nLocale.Get(i18n.I18nKeyPlatformBelow),
				},
				"gapSize": "large",
			},
			map[string]interface{}{
				"props": map[string]interface{}{
					"renderType": "Text",
					"visible":    visible,
					"value":      i18nLocale.Get(i18n.I18nKeyOrgSwitch),
					"styleConfig": map[string]interface{}{
						"bold": true,
					},
				},
				"gapSize": "small",
			},
			map[string]interface{}{
				"props": map[string]interface{}{
					"renderType": "Text",
					"visible":    visible,
					"value":      i18nLocale.Get(i18n.I18nKeyOrgSwitchMsg),
				},
				"gapSize": "large",
			},
			map[string]interface{}{
				"props": map[string]interface{}{
					"renderType": "Text",
					"visible":    visible,
					"value":      i18nLocale.Get(i18n.I18nKeyOrgPublicProBrowse),
					"styleConfig": map[string]interface{}{
						"bold": true,
					},
				},
				"gapSize": "small",
			},
			map[string]interface{}{
				"props": map[string]interface{}{
					"renderType": "Text",
					"visible":    visible,
					"value":      i18nLocale.Get(i18n.I18nKeyOrgLeftProBrowse),
				},
				"gapSize": "large",
			},
			map[string]interface{}{
				"props": map[string]interface{}{
					"renderType": "Text",
					"visible":    visible,
					"value":      i18nLocale.Get(i18n.I18nKeyOrgJoin),
					"styleConfig": map[string]interface{}{
						"bold": true,
					},
				},
				"gapSize": "small",
			},
			map[string]interface{}{
				"props": map[string]interface{}{
					"renderType": "Text",
					"visible":    visible,
					"value":      i18nLocale.Get(i18n.I18nKeyOrgJoinMsg),
				},
				"gapSize": "large",
			},
			map[string]interface{}{
				"props": map[string]interface{}{
					"renderType": "Text",
					"visible":    visible,
					"value":      i18nLocale.Get(i18n.I18nKeyOrgJoinAfter),
					"textStyleName": map[string]bool{
						"fz12":            true,
						"color-text-desc": true,
					},
				},
				"gapSize": "normal",
			},
		}
		return nil
	}
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
	if access {
		role = i18nLocale.Get(i18n.I18nKeyAdmin)
		createProType = "linkText"
		createProStr = i18nLocale.Get(i18n.I18nKeyProCreate)
		createProDetail = map[string]interface{}{
			"text": []interface{}{i18nLocale.Get(i18n.I18nKeyProCreateBy), map[string]interface{}{
				"icon":          "application-menu",
				"iconStyleName": "primary-icon",
			}, i18nLocale.Get(i18n.I18nKeyProCreateByCenter)},
		}
	} else {
		role = i18nLocale.Get(i18n.I18nKeyNewMember)
		createProType = "Text"
		createProStr = i18nLocale.Get(i18n.I18nKeyProJoin)
		createProDetail = i18nLocale.Get(i18n.I18nKeyProJoinMsg)
	}
	e.Props["visible"] = visible
	e.Props["value"] = []interface{}{
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "Text",
				"visible":    visible,
				"value":      fmt.Sprintf("%s%s%s", i18nLocale.Get(i18n.I18nKeyOrgBelow), role, i18nLocale.Get(i18n.I18nKeyOrgQuickKnow)),
			},
			"gapSize": "normal",
		},
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "Text",
				"visible":    visible,
				"value":      i18nLocale.Get(i18n.I18nKeyOrgSwitch),
				"styleConfig": map[string]interface{}{
					"bold": true,
				},
			},
			"gapSize": "small",
		},
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "Text",
				"visible":    visible,
				"value":      i18nLocale.Get(i18n.I18nKeyOrgSwitchMsg),
			},
			"gapSize": "large",
		},
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "Text",
				"visible":    visible,
				"value":      i18nLocale.Get(i18n.I18nKeyOrgBrowsePublic),
				"styleConfig": map[string]interface{}{
					"bold": true,
				},
			},
			"gapSize": "small",
		},
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "Text",
				"visible":    visible,
				"value":      i18nLocale.Get(i18n.I18nKeyOrgSwitchBrowse),
			},
			"gapSize": "large",
		},
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "Text",
				"visible":    visible,
				"value":      createProStr,
				"styleConfig": map[string]interface{}{
					"bold": true,
				},
			},
			"gapSize": "small",
		},
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": createProType,
				"visible":    visible,
				"value":      createProDetail,
			},
			"gapSize": "large",
		},
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "Text",
				"visible":    visible,
				"value":      i18nLocale.Get(i18n.I18nKeyOrgProBrowse),
				"styleConfig": map[string]interface{}{
					"bold": true,
				},
			},
			"gapSize": "small",
		},
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "linkText",
				"visible":    visible,
				"value": map[string]interface{}{
					"text": []interface{}{i18nLocale.Get(i18n.I18nKeyMenuClick), map[string]interface{}{
						"icon":          "application-menu",
						"iconStyleName": "primary-icon",
					}, i18nLocale.Get(i18n.I18nKeyProPublicBrowse)},
				},
			},
			"gapSize": "large",
		},
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "Text",
				"visible":    visible,
				"value":      i18nLocale.Get(i18n.I18nKeyProJoinAfter),
				"textStyleName": map[string]bool{
					"fz12":            true,
					"color-text-desc": true,
				},
			},
			"gapSize": "normal",
		},
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &EmptyProjectContent{}
}
