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

package emptyProject

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-sidebar/i18n"
)

func RenderCreator() protocol.CompRender {
	return &EmptyProject{}
}

type EmptyProject struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  props  `json:"props"`
	State  State  `json:"state"`
}

type props struct {
	Visible  bool   `json:"visible"`
	Tip      string `json:"tip"`
	Relative bool   `json:"relative"`
}

type State struct {
	ProsNum int `json:"prosNum"`
}

func (this *EmptyProject) GenComponentState(c *apistructs.Component) error {
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

func (this *EmptyProject) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *EmptyProject) setProps(visable bool, tip string) {
	this.Props.Visible = visable
	this.Props.Tip = tip
	this.Props.Relative = true
}

func (this *EmptyProject) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := this.GenComponentState(c); err != nil {
		return err
	}
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}
	prosNum := this.State.ProsNum
	i18nLocale := this.ctxBdl.Bdl.GetLocale(this.ctxBdl.Locale)
	if prosNum == 0 {
		this.setProps(true, i18nLocale.Get(i18n.I18nKeyProjectNoAdded))
		return nil
	} else {
		this.setProps(false, "")
	}
	return nil
}

func (this *EmptyProject) getProjectsNum(orgID string) (int, error) {
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
