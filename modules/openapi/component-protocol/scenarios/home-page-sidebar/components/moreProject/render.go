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

package moreProject

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func RenderCreator() protocol.CompRender {
	return &MoreProject{}
}

type MoreProject struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  Props  `json:"props"`
	State  State  `json:"state"`
}

type Props struct {
	RenderType string `json:"renderType"`
	Visible    bool   `json:"visible"`
	Value      Value  `json:"value"`
}

type Value struct {
	Text string `json:`
}

type State struct {
	//HavePros bool `json:"havePros"`
	//HaveApps bool `json:"haveApps"`
	//OrgID string `json:"orgID"`
}

// GenComponentState 获取state
func (this *MoreProject) GenComponentState(c *apistructs.Component) error {
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

func (this *MoreProject) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *MoreProject) setComponentValue() {
	this.Type = "Text"
	this.Props.RenderType = "linkText"
	this.Props.Visible = true
	this.Props.Value.Text = "更多"
}

func (this *MoreProject) getProjectsNum(orgID string) (int, error) {
	orgIntId, err := strconv.Atoi(orgID)
	if err != nil {
		return 0, err
	}
	req := apistructs.ProjectListRequest{
		OrgID:    uint64(orgIntId),
		PageNo:   1,
		PageSize: 1,
		IsPublic: true,
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

func (this *MoreProject) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	//if err := this.GenComponentState(c); err != nil {
	//	return err
	//}
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}
	//if !this.State.HavePros {
	//	this.Props.Visible = false
	//	return nil
	//}
	prosNum, err := this.getProjectsNum(this.ctxBdl.Identity.OrgID)
	if err != nil {
		return nil
	}
	if prosNum == 0 {
		this.Props.Visible = false
		return nil
	}
	this.setComponentValue()
	return nil
}
