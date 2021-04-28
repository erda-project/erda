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

package emptyProjectTitle

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type EmptyProjectTitle struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  Props  `json:"props"`
	State  State  `json:"state"`
}

type Props struct {
	Visible bool   `json:"visible"`
	Title   string `json:"title"`
	Level   int    `json:"level"`
}

type State struct {
	ProsNum int `json:"prosNum"`
}

func (this *EmptyProjectTitle) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *EmptyProjectTitle) GenComponentState(c *apistructs.Component) error {
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

func (this *EmptyProjectTitle) getProjectsNum(orgID string) (int, error) {
	orgIntId, err := strconv.Atoi(orgID)
	if err != nil {
		return 0, err
	}
	req := apistructs.ProjectListRequest{
		OrgID:    uint64(orgIntId),
		PageNo:   1,
		PageSize: 1,
	}

	projectDTO, err := this.ctxBdl.Bdl.ListMyProject(this.ctxBdl.Identity.UserID, req)
	if err != nil {
		return 0, err
	}
	if projectDTO == nil {
		return 0, nil
	}
	return projectDTO.Total, nil
}

func (e *EmptyProjectTitle) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := e.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := e.GenComponentState(c); err != nil {
		return err
	}
	e.Type = "Title"
	e.Props.Title = "你已经是 Erda Cloud 组织的成员"
	e.Props.Level = 2
	if e.ctxBdl.Identity.OrgID != "" {
		prosNum, err := e.getProjectsNum(e.ctxBdl.Identity.OrgID)
		if err != nil {
			return err
		}
		if prosNum == 0 {
			orgDTO, err := e.ctxBdl.Bdl.GetOrg(e.ctxBdl.Identity.OrgID)
			if err != nil {
				return err
			}
			if orgDTO == nil {
				return fmt.Errorf("can not get org")
			}
			orgIntId, err := strconv.Atoi(e.ctxBdl.Identity.OrgID)
			if err != nil {
				return err
			}
			req := &apistructs.PermissionCheckRequest{
				UserID:   e.ctxBdl.Identity.UserID,
				Scope:    apistructs.OrgScope,
				ScopeID:  uint64(orgIntId),
				Resource: apistructs.ProjectResource,
				Action:   apistructs.CreateAction,
			}
			var role string = "成员"
			permissionRes, err := e.ctxBdl.Bdl.CheckPermission(req)
			if err != nil {
				return err
			}
			if permissionRes == nil {
				return fmt.Errorf("can not check permission for org")
			}

			if permissionRes.Access {
				role = "管理员"
			}
			e.Props.Title = fmt.Sprintf("你已经是 %s 组织的%s", orgDTO.DisplayName, role)
			e.Props.Visible = true
		}
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &EmptyProjectTitle{}
}
