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

package orgSwitch

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-sidebar/i18n"
)

const (
	DefaultPageSize = 100
)

func RenderCreator() protocol.CompRender {
	return &OrgSwitch{}
}

type OrgSwitch struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  props  `json:"props"`
	//Data Data `json:"data"`
	State State `json:"state"`
}

type Data struct {
	List []OrgItem `json:"list"`
}

type OrgItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	IsPublic    bool   `json:"is_public"`
	Status      string `json:"status"`
	Logo        string `json:"logo"`
}

type props struct {
	Visible     bool          `json:"visible"`
	Options     []MenuItem    `json:"options"`
	QuickSelect []interface{} `json:"quickSelect"`
}

type MenuItem struct {
	Label        string                 `json:"label"`
	Value        string                 `json:"value"`
	PrefixImgSrc string                 `json:"prefixImgSrc"`
	Operations   map[string]interface{} `json:"operations"`
}

type Meta struct {
	Id       string `json:"id"`
	Severity string `json:"severity"`
}

type State struct {
	//OrgID string `json:"orgID"`
	Value string `json:"value"`
	Label string `json:"label"`
}

type Operation struct {
	Key        string `json:"key"`
	Reload     bool   `json:"reload"`
	Disabled   bool   `json:"disabled"`
	Text       string `json:"text"`
	PrefixIcon string `json:"prefixIcon"`
	Meta       Meta   `json:"meta"`
}

func (this *OrgSwitch) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *OrgSwitch) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}
	if this.ctxBdl.Identity.OrgID == "" {
		this.Props.Visible = false
		return nil
	}
	this.Type = "DropdownSelect"
	this.Props.Visible = true
	//if err := this.setComponentValue(""); err != nil {
	//	return err
	//}
	if err := this.RenderList(); err != nil {
		return err
	}
	orgDTO, err := this.ctxBdl.Bdl.GetOrg(this.ctxBdl.Identity.OrgID)
	if err != nil {
		return err
	}
	if orgDTO == nil {
		return fmt.Errorf("can not get org")
	}
	i18nLocale := this.ctxBdl.Bdl.GetLocale(this.ctxBdl.Locale)
	this.State.Value = strconv.FormatInt(int64(orgDTO.ID), 10)
	this.State.Label = orgDTO.DisplayName
	this.Props.QuickSelect = []interface{}{
		map[string]interface{}{
			"value": "orgList",
			"label": i18nLocale.Get(i18n.I18nKeyOrgBrowse),
			"operations": map[string]interface{}{
				"click": map[string]interface{}{
					"key":    "click",
					"show":   false,
					"reload": false,
					"command": map[string]interface{}{
						"key":     "goto",
						"target":  "orgList",
						"jumpOut": false,
					},
				},
			},
		},
	}
	return nil
}

func RenItem(org apistructs.OrgDTO) MenuItem {
	logo := "//terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2021/06/03/9b1a8af7-0111-4c14-9158-9804bb3ebafc.png"
	if org.Logo != "" {
		logo = org.Logo
	}
	item := MenuItem{
		Label:        org.DisplayName,
		Value:        strconv.FormatInt(int64(org.ID), 10),
		PrefixImgSrc: logo,
		Operations: map[string]interface{}{
			"click": map[string]interface{}{
				"key":    "click",
				"show":   false,
				"reload": false,
				"command": map[string]interface{}{
					"key":     "goto",
					"target":  "orgRoot",
					"jumpOut": false,
					"state": map[string]interface{}{
						"params": map[string]interface{}{
							"orgName": org.Name,
						},
					},
				},
			},
		},
	}
	return item
}

func (this *OrgSwitch) RenderList() error {
	identity := apistructs.IdentityInfo{UserID: this.ctxBdl.Identity.UserID}
	req := &apistructs.OrgSearchRequest{
		IdentityInfo: identity,
		PageSize:     DefaultPageSize,
	}
	pagingOrgDTO, err := this.ctxBdl.Bdl.ListDopOrgs(req)
	if err != nil {
		return err
	}
	this.Props.Options = make([]MenuItem, 0)
	for _, v := range pagingOrgDTO.List {
		this.Props.Options = append(this.Props.Options, RenItem(v))
	}
	return nil
}
