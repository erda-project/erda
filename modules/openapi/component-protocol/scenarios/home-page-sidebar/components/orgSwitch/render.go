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

package orgSwitch

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
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
	this.State.Value = strconv.FormatInt(int64(orgDTO.ID), 10)
	this.Props.QuickSelect = []interface{}{
		map[string]interface{}{
			"value": "orgList",
			"label": "浏览公开组织",
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
	logo := "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQYQY0vUTJwftJ8WqXoLiLeB--2MJkpZLpYOA&usqp=CAU"
	if org.Logo != "" {
		logo = fmt.Sprintf("https:%s", org.Logo)
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
	pagingOrgDTO, err := this.ctxBdl.Bdl.ListOrgs(req)
	if err != nil {
		return err
	}
	this.Props.Options = make([]MenuItem, 0)
	for _, v := range pagingOrgDTO.List {
		this.Props.Options = append(this.Props.Options, RenItem(v))
	}
	return nil
}
