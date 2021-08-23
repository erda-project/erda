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

package myApplicationFilter

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
	return &MyApplicationFilter{}
}

type MyApplicationFilter struct {
	ctxBdl     protocol.ContextBundle
	Type       string                 `json:"type"`
	Props      Props                  `json:"props"`
	State      State                  `json:"state"`
	Operations map[string]interface{} `json:"operations"`
}

type Props struct {
	Visible   bool `json:"visible"`
	Delay     int  `json:"delay"`
	FullWidth bool `json:"fullWidth"`
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
	//HaveApps bool `json:"haveApps"`
	IsFirstFilter bool                   `json:"isFirstFilter"`
	Values        map[string]interface{} `json:"values"`
	ProsNum       int                    `json:"prosNum"`
	AppsNum       int                    `json:"appsNum"`
	//OrgID string `json:"orgID"`
}

type Operation struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}

// GenComponentState 获取state
func (this *MyApplicationFilter) GenComponentState(c *apistructs.Component) error {
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

func (this *MyApplicationFilter) setComponentValue() {
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
			Placeholder: i18nLocale.Get(i18n.I18nKeyFilterSearchApp),
			Type:        "input",
		},
	}
}

func (this *MyApplicationFilter) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *MyApplicationFilter) getAppsNum(orgID string) (int, error) {
	orgIntId, err := strconv.Atoi(orgID)
	if err != nil {
		return 0, err
	}
	req := apistructs.ApplicationListRequest{PageSize: 1, PageNo: 1}
	appsDTO, err := this.ctxBdl.Bdl.GetAllMyApps(this.ctxBdl.Identity.UserID, uint64(orgIntId), req)
	if err != nil {
		return 0, err
	}
	if appsDTO == nil {
		return 0, nil
	}
	return appsDTO.Total, nil
}

func (this *MyApplicationFilter) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := this.GenComponentState(c); err != nil {
		return err
	}
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}
	//if !this.State.HaveApps {
	//	this.Props.Visible = false
	//	return nil
	//}
	if this.State.AppsNum == 0 {
		this.Props.Visible = false
		return nil
	}
	this.setComponentValue()

	this.Props.Visible = true
	this.State.IsFirstFilter = false
	if event.Operation == apistructs.ListProjectFilterOperation {
		this.State.IsFirstFilter = true
	}

	return nil
}
