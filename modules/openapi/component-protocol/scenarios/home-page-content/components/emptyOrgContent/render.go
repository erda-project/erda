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
	if access {
		role = "管理员"
	} else {
		role = "新成员"
	}
	e.Props.Value = make([]Value, 0)
	e.Props.Value = append(e.Props.Value, Value{
		Props: PropValue{
			RenderType: "text",
			Visible:    visible,
			Value:      fmt.Sprintf("以下是作为组织%s的一些快速入门知识：", role),
		},
		GapSize: "large",
	}, Value{
		Props: PropValue{
			RenderType: "text",
			Visible:    visible,
			Value:      "* 公开组织浏览",
			StyleConfig: map[string]bool{
				"bold": true,
			},
		},
		GapSize: "small",
	}, Value{
		Props: PropValue{
			RenderType: "text",
			Visible:    visible,
			Value:      "通过左上角的浏览公开组织信息，选择公开组织可以直接进入浏览该组织公开项目的信息可（包含项目管理、应用运行信息等）",
		},
		GapSize: "large",
	}, Value{
		Props: PropValue{
			RenderType: "text",
			Visible:    visible,
			Value:      "* 加入组织",
			StyleConfig: map[string]bool{
				"bold": true,
			},
		},
		GapSize: "small",
	}, Value{
		Props: PropValue{
			RenderType: "text",
			Visible:    visible,
			Value:      "组织当前都是受邀机制，需要线下联系企业所有者进行邀请加入",
		},
		GapSize: "large",
	}, Value{
		Props: PropValue{
			RenderType: "text",
			Visible:    visible,
			Value:      "当你已经加入到任何组织后，此框将不再显示",
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
