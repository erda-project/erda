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

package joinedBrief

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

const (
	DefaultType = "Table"
)

func RenderCreator() protocol.CompRender {
	return &JoinedBrief{}
}

type JoinedBrief struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  props  `json:"props"`
	Data   data   `json:"data"`
	State  State  `json:"state"`
}

type State struct {
	//OrgID string `json:"orgID"`
	//JoinedOrg bool `json:"joinedOrg"`
	//HavePros bool `json:"havePros"`
	//HaveApps bool `json:"haveApps"`
	ProsNum int `json:"prosNum"`
}

type column struct {
	Title     string `json:"title"`
	DataIndex string `json:"dataIndex"`
	Width     int    `json:"width,omitempty"`
}

type props struct {
	Visible    bool       `json:"visible"`
	RowKey     string     `json:"rowKey"`
	Columns    []column   `json:"columns"`
	ShowHeader bool       `json:"showHeader"`
	Pagination bool       `json:"pagination"`
	StyleNames StyleNames `json:"styleNames"`
}

type StyleNames struct {
	NoBorder  bool `json:"no-border"`
	LightCard bool `json:"light-card"`
}

type category struct {
	RenderType     string `json:"renderType"`
	PrefixIcon     string `json:"prefixIcon"`
	Value          string `json:"value"`
	ColorClassName string `json:"colorClassName"`
}

type bItem struct {
	Id       int      `json:"id"`
	Category category `json:"category"`
	Number   int      `json:"number"`
}

type data struct {
	List []bItem `json:"list"`
}

func (this *JoinedBrief) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

// GenComponentState 获取state
func (this *JoinedBrief) GenComponentState(c *apistructs.Component) error {
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

func (this *JoinedBrief) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := this.GenComponentState(c); err != nil {
		return err
	}
	this.Type = "Table"
	orgID := this.ctxBdl.Identity.OrgID
	if this.ctxBdl.Identity.OrgID != "" {
		this.Props.Visible = true
		if err := this.setData(orgID); err != nil {
			return err
		}
	}
	this.setProps()
	return nil
}

func (this *JoinedBrief) setProps() {
	this.Props.Columns = make([]column, 0)
	this.Props.RowKey = "key"
	this.Props.ShowHeader = false
	this.Props.Pagination = false

	this.Props.Columns = append(this.Props.Columns, column{Title: "", DataIndex: "category"})
	this.Props.Columns = append(this.Props.Columns, column{Title: "", DataIndex: "number", Width: 42})
	this.Props.StyleNames = StyleNames{
		NoBorder:  true,
		LightCard: true,
	}
}

func (this *JoinedBrief) setData(orgID string) error {
	this.Data.List = make([]bItem, 0)
	projectNum, err := this.getProjectsNum(orgID)
	if err != nil {
		return err
	}
	//if projectNum > 0 {
	//	this.State.HavePros = true
	//}
	this.Data.List = append(this.Data.List, bItem{
		Id: 1, Category: category{
			RenderType:     "textWithIcon",
			PrefixIcon:     "project-icon",
			Value:          "参与项目数：",
			ColorClassName: "color-primary"},
		Number: projectNum})
	appNum, err := this.getAppsNum(orgID)
	if err != nil {
		return err
	}
	//if appNum > 0 {
	//	this.State.HaveApps = true
	//}
	this.Data.List = append(this.Data.List, bItem{
		Id: 1, Category: category{
			RenderType:     "textWithIcon",
			PrefixIcon:     "app-icon",
			Value:          "参与应用数：",
			ColorClassName: "color-primary"},
		Number: appNum})
	return nil
}

func (this *JoinedBrief) getProjectsNum(orgID string) (int, error) {
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

func (this *JoinedBrief) getAppsNum(orgID string) (int, error) {
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
