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

package myProjectFilter

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
	return &MyProjectFilter{}
}

type MyProjectFilter struct {
	ctxBdl     protocol.ContextBundle
	Type       string                 `json:"type"`
	Props      Props                  `json:"props"`
	State      State                  `json:"state"`
	Operations map[string]interface{} `json:"operations"`
}

type Props struct {
	FullWidth bool `json:"fullWidth"`
	Visible   bool `json:"visible"`
	Delay     int  `json:"delay"`
}

type Condition struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	EmptyText   string `json:"emptyText"`
	Fixed       bool   `json:"fixed"`
	ShowIndex   int    `json:"showIndex"`
	Placeholder string `json:"placeholder"`
	Type        string `json:"type"`
}

type State struct {
	Conditions []Condition `json:"conditions"`
	//HavePros bool `json:"havePros"`
	//HaveApps bool `json:"haveApps"`
	IsFirstFilter bool                   `json:"isFirstFilter"`
	Values        map[string]interface{} `json:"values"`
	ProsNum       int                    `json:"prosNum"`
	//OrgID         string                 `json:"orgID"`
}

type Operation struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}

// GenComponentState 获取state
func (this *MyProjectFilter) GenComponentState(c *apistructs.Component) error {
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

func (this *MyProjectFilter) setComponentValue() {
	i18nLocale := this.ctxBdl.Bdl.GetLocale(this.ctxBdl.Locale)
	this.Props.Delay = 1000
	this.Props.FullWidth = true
	this.Operations = map[string]interface{}{
		apistructs.ListProjectFilterOperation.String(): Operation{
			Reload: true,
			Key:    apistructs.ListProjectFilterOperation.String(),
		},
	}
	this.State.Conditions = []Condition{
		{
			Key:         "title",
			Label:       i18nLocale.Get(i18n.I18nKeyFilterTitle),
			EmptyText:   i18nLocale.Get(i18n.I18nKeyAll),
			Fixed:       true,
			ShowIndex:   2,
			Placeholder: i18nLocale.Get(i18n.I18nKeyFilterSearchPro),
			Type:        "input",
		},
	}
}

// RenderProtocol 渲染组件
func (this *MyProjectFilter) RenderProtocol(c *apistructs.Component, g *apistructs.GlobalStateData) error {
	stateValue, err := json.Marshal(this.State)
	if err != nil {
		return err
	}
	var state map[string]interface{}
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}
	c.State = state

	c.Props = this.Props
	c.Operations = this.Operations
	return nil
}

func (this *MyProjectFilter) getProjectsNum(orgID string) (int, error) {
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

func (this *MyProjectFilter) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *MyProjectFilter) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := this.GenComponentState(c); err != nil {
		return err
	}
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}

	//prosNum, err := this.getProjectsNum(this.ctxBdl.Identity.OrgID)
	//if err != nil {
	//	return err
	//}
	//if prosNum == 0 {
	//	this.Props.Visible = false
	//	return nil
	//}
	if this.ctxBdl.Identity.OrgID == "" {
		return nil
	}
	if this.State.ProsNum != 0 {
		this.Props.Visible = true
	}
	//this.State.IsFirstFilter = false
	//if event.GenerateOperation == apistructs.ListProjectFilterOperation {
	//	this.State.IsFirstFilter = true
	//}
	this.setComponentValue()

	//if err := this.RenderProtocol(c, gs); err != nil {
	//	return err
	//}
	return nil
}
