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
	var visible bool
	if e.ctxBdl.Identity.OrgID != "" && e.State.ProsNum == 0 {
		visible = true
	}

	var access bool
	var role string
	var createProStr string
	var createProDetail interface{}
	var createProType string
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
		createProType = "linkText"
		createProStr = "* 创建项目"
		createProDetail = map[string]interface{}{
			"text": []interface{}{"通过左上角菜单快读创建或者进入", map[string]interface{}{
				"icon":          "appstore",
				"iconStyleName": "primary-icon",
			}, "选择管理中心后, 进行项目的创建"},
		}
	} else {
		role = "新成员"
		createProType = "Text"
		createProStr = "* 加入项目"
		createProDetail = "当前都是受邀机制，需要线下联系项目管理员进行邀请加入"
	}
	e.Props = make(map[string]interface{})
	e.Props["visible"] = visible
	e.Props["value"] = []interface{}{
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "Text",
				"visible":    visible,
				"value":      fmt.Sprintf("以下是作为组织%s的一些快速入门知识：", role),
			},
			"gapSize": "normal",
		},
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "Text",
				"visible":    visible,
				"value":      "* 切换组织",
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
				"value":      "使用此屏幕上左上角的组织切换，快速进行组织之间切换",
			},
			"gapSize": "large",
		},
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "Text",
				"visible":    visible,
				"value":      "* 公开组织浏览",
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
				"value":      "可以通过切换组织下拉菜单中选择公开组织进行浏览",
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
				"value":      "* 该组织内公开项目浏览",
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
					"text": []interface{}{"点击左上角菜单", map[string]interface{}{
						"icon":          "application-menu",
						"iconStyleName": "primary-icon",
					}, "选择 DevOps平台进入，选择我的项目可以查看该组织下公开项目的信息"},
				},
			},
			"gapSize": "large",
		},
		map[string]interface{}{
			"props": map[string]interface{}{
				"renderType": "Text",
				"visible":    visible,
				"value":      "当你已经加入到任何项目后，此框将不再显示",
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
