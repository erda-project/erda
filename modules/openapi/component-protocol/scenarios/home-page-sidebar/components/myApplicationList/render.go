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

package myApplicationList

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
	DefaultPageNo   = 1
	DefaultPageSize = 5
)

type MyApplicationList struct {
	ctxBdl     protocol.ContextBundle
	Type       string                 `json:"type"`
	Props      Props                  `json:"props"`
	State      State                  `json:"state"`
	Data       Data                   `json:"data"`
	Operations map[string]interface{} `json:"operations"`
}

type AppItem struct {
	ID          string               `json:"id"`
	AppId       string               `json:"appId"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	PrefixImg   string               `json:"prefixImg"`
	Operations  map[string]Operation `json:"operations"`
}

type OperationData struct {
	FillMeta string `json:"fillMeta"`
	Meta     Meta   `json:"meta"`
}

type Meta struct {
	PageNo PageNo `json:"pageNo"`
}

type PageNo struct {
	PageNo int `json:"pageNo"`
}

type Command struct {
	Key    string                 `json:"key"`
	Target string                 `json:"target"`
	State  map[string]interface{} `json:"state"`
}

type Operation struct {
	Command Command `json:"command"`
	Key     string  `json:"key"`
	Reload  bool    `json:"reload"`
	Show    bool    `json:"show"`
}

type Data struct {
	List []AppItem `json:"list"`
}

func (s Data) Less(i, j int) bool {
	return s.List[i].Title < s.List[j].Title
}

func (s Data) Swap(i, j int) {
	s.List[i], s.List[j] = s.List[j], s.List[i]
}

func (s Data) Len() int {
	return len(s.List)
}

type Props struct {
	Visible     bool   `json:"visible"`
	IsLoadMore  bool   `json:"isLoadMore"`
	AlignCenter bool   `json:"alignCenter"`
	Size        string `json:"size"`
	NoBorder    bool   `json:"noBorder"`
}

type State struct {
	//HaveApps bool `json:"haveApps"`
	IsFirstFilter bool                   `json:"isFirstFilter"`
	Values        map[string]interface{} `json:"values"`
	PageNo        int                    `json:"pageNo"`
	PageSize      int                    `json:"pageSize"`
	Total         int                    `json:"total"`
	ProsNum       int                    `json:"prosNum"`
	AppsNum       int                    `json:"appsNum"`
	//OrgID string `json:"orgID"`
}

func (this *MyApplicationList) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

// GenComponentState 获取state
func (this *MyApplicationList) GenComponentState(c *apistructs.Component) error {
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

func RenItem(app apistructs.ApplicationDTO, orgName string) AppItem {
	item := AppItem{
		ID:          strconv.Itoa(int(app.ID)),
		AppId:       strconv.Itoa(int(app.ID)),
		Title:       fmt.Sprintf("%s/%s", app.ProjectDisplayName, app.DisplayName),
		Description: "",
		PrefixImg:   "frontImg_default_app_icon",
		Operations: map[string]Operation{
			"click": {
				Key:    "click",
				Show:   false,
				Reload: false,
				Command: Command{
					Key:    "goto",
					Target: "app",
					State: map[string]interface{}{
						"params": map[string]interface{}{
							"projectId": strconv.Itoa(int(app.ProjectID)),
							"appId":     strconv.Itoa(int(app.ID)),
							"orgName":   orgName,
						},
					},
				},
			},
		},
	}
	if app.Logo != "" {
		item.PrefixImg = app.Logo
	}
	return item
}

func (this *MyApplicationList) getAppsNum(orgID string, queryStr string) (int, error) {
	orgIntId, err := strconv.Atoi(orgID)
	if err != nil {
		return 0, err
	}
	req := apistructs.ApplicationListRequest{
		PageSize: 1,
		PageNo:   1,
		Query:    queryStr,
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

func (m *MyApplicationList) getAppDTO(orgID string, queryStr string) (*apistructs.ApplicationListResponseData, error) {
	orgIDInt, err := strconv.ParseUint(m.ctxBdl.Identity.OrgID, 10, 64)
	if err != nil {
		return nil, err
	}
	req := apistructs.ApplicationListRequest{
		PageSize: m.State.PageSize,
		PageNo:   m.State.PageNo,
		Query:    queryStr,
		OrderBy:  "name",
		IsSimple: true,
	}
	appsDTO, err := m.ctxBdl.Bdl.GetAllMyApps(m.ctxBdl.Identity.UserID, uint64(orgIDInt), req)
	if err != nil {
		return nil, err
	}
	if appsDTO == nil {
		return nil, nil
	}
	return appsDTO, nil
}

func (m *MyApplicationList) addAppsData(datas *apistructs.ApplicationListResponseData) error {
	var orgName string
	dataList := make([]AppItem, 0)
	if len(datas.List) > 0 {
		orgDTO, err := m.ctxBdl.Bdl.GetOrg(m.ctxBdl.Identity.OrgID)
		if err != nil {
			return err
		}
		if orgDTO == nil {
			return fmt.Errorf("failed to get org")
		}
		orgName = orgDTO.Name
	}
	for _, v := range datas.List {
		dataList = append(dataList, RenItem(v, orgName))
	}
	m.Data.List = dataList
	return nil
}

func (this *MyApplicationList) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	//if !this.State.HaveApps {
	//	this.Props.Visible = false
	//	return nil
	//}
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := this.GenComponentState(c); err != nil {
		return err
	}
	if this.State.AppsNum == 0 {
		this.Props.Visible = false
		return nil
	}
	this.Props.Visible = true
	this.Props.IsLoadMore = true
	this.Props.AlignCenter = true
	this.Props.Size = "small"
	this.Props.NoBorder = true

	queryStr := ""
	_, ok := this.State.Values["title"]
	if ok {
		queryStr = this.State.Values["title"].(string)
	}

	switch event.Operation {
	case apistructs.InitializeOperation:
		this.Props.Visible = true
		this.State.PageNo = DefaultPageNo
		this.State.PageSize = DefaultPageSize
		appsDTO, err := this.getAppDTO(this.ctxBdl.Identity.OrgID, queryStr)
		if err != nil {
			return err
		}
		this.State.Total = 0
		if appsDTO != nil {
			if err := this.addAppsData(appsDTO); err != nil {
				return err
			}
			this.State.Total = appsDTO.Total
		}
		this.Operations = map[string]interface{}{
			"changePageNo": map[string]interface{}{
				"key":      "changePageNo",
				"reload":   true,
				"fillMeta": "pageNo",
			},
		}
	case apistructs.RenderingOperation:
		this.Data.List = make([]AppItem, 0)
		this.State.PageNo = DefaultPageNo
		this.State.PageSize = DefaultPageSize
		appsDTO, err := this.getAppDTO(this.ctxBdl.Identity.OrgID, queryStr)
		if err != nil {
			return err
		}
		this.State.Total = 0
		if appsDTO != nil {
			if err := this.addAppsData(appsDTO); err != nil {
				return err
			}
			this.State.Total = appsDTO.Total
		}
	case apistructs.OnChangePageNoOperation:
		var pageData OperationData
		dataBody, err := json.Marshal(event.OperationData)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(dataBody, &pageData); err != nil {
			return err
		}
		this.State.PageNo = pageData.Meta.PageNo.PageNo
		appsDTO, err := this.getAppDTO(this.ctxBdl.Identity.OrgID, queryStr)
		if err != nil {
			return err
		}
		this.State.Total = 0
		if appsDTO != nil {
			if err := this.addAppsData(appsDTO); err != nil {
				return err
			}
			this.State.Total = appsDTO.Total
		}
	}
	//sort.Sort(this.Data)
	return nil
}

func RenderCreator() protocol.CompRender {
	return &MyApplicationList{}
}
