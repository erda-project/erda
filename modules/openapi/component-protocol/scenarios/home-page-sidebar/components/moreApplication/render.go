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

package moreApplication

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
	return &MoreApplication{}
}

type MoreApplication struct {
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
	//HaveApps bool `json:"haveApps"`
}

// GenComponentState 获取state
func (this *MoreApplication) GenComponentState(c *apistructs.Component) error {
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

func (this *MoreApplication) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *MoreApplication) setComponentValue() {
	i18nLocale := this.ctxBdl.Bdl.GetLocale(this.ctxBdl.Locale)
	this.Type = "Text"
	this.Props.RenderType = "linkText"
	this.Props.Visible = true
	this.Props.Value.Text = i18nLocale.Get(i18n.I18nKeyMore)
}

func (this *MoreApplication) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	//if err := this.GenComponentState(c); err != nil {
	//	return err
	//}
	//if !this.State.HaveApps {
	//	this.Props.Visible = false
	//	return nil
	//}
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}

	appsNum, err := this.getAppsNum(this.ctxBdl.Identity.OrgID)
	if err != nil {
		return err
	}
	if appsNum == 0 {
		this.Props.Visible = false
		return nil
	}
	this.setComponentValue()
	return nil
}

func (this *MoreApplication) getAppsNum(orgID string) (int, error) {
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
