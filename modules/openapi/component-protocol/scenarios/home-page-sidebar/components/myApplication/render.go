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

package myApplication

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type MyApplication struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  Props  `json:"props"`
	State  State  `json:"state"`
}

type Props struct {
	Visible bool `json:"visible"`
}

type State struct {
	ProsNum int `json:"prosNum"`
	AppsNum int `json:"appsNum"`
}

func (this *MyApplication) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *MyApplication) getAppsNum(orgID string, queryStr string) (int, error) {
	orgIntId, err := strconv.Atoi(orgID)
	if err != nil {
		return 0, err
	}
	req := apistructs.ApplicationListRequest{
		PageSize: 1,
		PageNo:   1,
		Query:    queryStr,
		IsSimple: true,
	}
	appsDTO, err := this.ctxBdl.Bdl.GetAllMyApps(this.ctxBdl.Identity.UserID, uint64(orgIntId), req)
	if err != nil {
		return 0, err
	}
	if appsDTO == nil {
		return 0, nil
	}
	return appsDTO.Total, nil
}

func (this *MyApplication) GenComponentState(c *apistructs.Component) error {
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

func (this *MyApplication) getProjectsNum(orgID string) (int, error) {
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

func (t *MyApplication) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := t.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := t.GenComponentState(c); err != nil {
		return err
	}
	var visible bool
	if t.ctxBdl.Identity.OrgID == "" || t.State.ProsNum == 0 {
		visible = false
	} else {
		appsNum, err := t.getAppsNum(t.ctxBdl.Identity.OrgID, "")
		if err != nil {
			return err
		}
		t.State.AppsNum = appsNum
		visible = true
	}

	t.Type = "Container"
	t.Props.Visible = visible
	return nil
}

func RenderCreator() protocol.CompRender {
	return &MyApplication{}
}
