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

package createProjectTip

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-sidebar/i18n"
	"github.com/erda-project/erda/pkg/strutil"
)

type CreateProjectTip struct {
	ctxBdl     protocol.ContextBundle
	Type       string               `json:"type"`
	Props      Props                `json:"props"`
	Operations map[string]Operation `json:"operations"`
	State      State                `json:"state"`
}

type Props struct {
	Visible    bool                   `json:"visible"`
	RenderType string                 `json:"renderType"`
	Value      map[string]interface{} `json:"value"`
}

type State struct {
	ProsNum int `json:"prosNum"`
}

type Command struct {
	Key     string `json:"key"`
	Target  string `json:"target"`
	JumpOut bool   `json:"jumpOut"`
	Visible bool   `json:"visible"`
}

type Operation struct {
	Command Command `json:"command"`
	Key     string  `json:"key"`
	Reload  bool    `json:"reload"`
	Show    bool    `json:"show"`
}

func (this *CreateProjectTip) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *CreateProjectTip) GenComponentState(c *apistructs.Component) error {
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

func (this *CreateProjectTip) getProjectsNum(orgID string) (int, error) {
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

func (p *CreateProjectTip) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := p.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := p.GenComponentState(c); err != nil {
		return err
	}
	var visible bool
	if p.ctxBdl.Identity.OrgID != "" && p.State.ProsNum == 0 {
		visible = true
	}

	i18nLocale := p.ctxBdl.Bdl.GetLocale(p.ctxBdl.Locale)
	p.Type = "Text"
	p.Props.Visible = visible
	p.Props.RenderType = "linkText"
	p.Props.Value = map[string]interface{}{
		"text": []interface{}{map[string]interface{}{
			"text":         i18nLocale.Get(i18n.I18nKeyProjectHowCreate),
			"operationKey": "createProjectDoc",
		}, fmt.Sprintf(" %s ", i18nLocale.Get(i18n.I18nKeyOr)), map[string]interface{}{
			"text":         i18nLocale.Get(i18n.I18nKeyOrgBrowseInfo),
			"operationKey": "toPublicOrgPage",
		}},
	}
	p.Operations = map[string]Operation{
		"createProjectDoc": {
			Command: Command{
				Key:     "goto",
				Target:  strutil.Concat("https://docs.erda.cloud/", version.Version, "/manual/platform-design.html#%E9%A1%B9%E7%9B%AE%E5%92%8C%E5%BA%94%E7%94%A8"),
				JumpOut: true,
				Visible: visible,
			},
			Key:    "click",
			Reload: false,
			Show:   false,
		},
		"toPublicOrgPage": {
			Command: Command{
				Key:     "goto",
				Target:  "orgList",
				JumpOut: true,
				Visible: visible,
			},
			Key:    "click",
			Reload: false,
			Show:   false,
		},
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &CreateProjectTip{}
}
