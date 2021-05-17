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

package myProjectList

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

type MyProjectList struct {
	ctxBdl     protocol.ContextBundle
	Type       string                 `json:"type"`
	Props      Props                  `json:"props"`
	State      State                  `json:"state"`
	Data       Data                   `json:"data"`
	Operations map[string]interface{} `json:"operations"`
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

type ProItem struct {
	ID          string               `json:"id"`
	ProjectId   string               `json:"projectId"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	PrefixImg   string               `json:"prefixImg"`
	Operations  map[string]Operation `json:"operations"`
}

type Data struct {
	List []ProItem `json:"list"`
}

type Props struct {
	Visible     bool   `json:"visible"`
	UseLoadMore bool   `json:"useLoadMore"`
	AlignCenter bool   `json:"alignCenter"`
	Size        string `json:"size"`
	NoBorder    bool   `json:"noBorder"`
	//PaginationType string `json:"paginationType"`
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

func (s Data) Less(i, j int) bool {
	return s.List[i].Title < s.List[j].Title
}

func (s Data) Swap(i, j int) {
	s.List[i], s.List[j] = s.List[j], s.List[i]
}

func (s Data) Len() int {
	return len(s.List)
}

type State struct {
	//HavePros bool `json:"havePros"`
	//HaveApps bool `json:"haveApps"`
	IsFirstFilter bool                   `json:"isFirstFilter"`
	Values        map[string]interface{} `json:"values"`
	PageNo        int                    `json:"pageNo"`
	PageSize      int                    `json:"pageSize"`
	Total         int                    `json:"total"`
	ProNums       int                    `json:"prosNum"`
	//OrgID         string                 `json:"orgID"`
}

func (this *MyProjectList) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

// GenComponentState 获取state
func (this *MyProjectList) GenComponentState(c *apistructs.Component) error {
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

func RenItem(pro apistructs.ProjectDTO, orgName string) ProItem {
	item := ProItem{
		ID:          strconv.Itoa(int(pro.ID)),
		ProjectId:   strconv.Itoa(int(pro.ID)),
		Title:       pro.DisplayName,
		Description: "",
		PrefixImg:   "/images/default-project-icon.png",
		Operations: map[string]Operation{
			"click": {
				Key:    "click",
				Show:   false,
				Reload: false,
				Command: Command{
					Key:    "goto",
					Target: "projectAllIssue",
					State: map[string]interface{}{
						"params": map[string]interface{}{
							"projectId": strconv.Itoa(int(pro.ID)),
							"orgName":   orgName,
						},
					},
				},
			},
		},
	}
	if pro.Logo != "" {
		item.PrefixImg = fmt.Sprintf("https:%s", pro.Logo)
	}
	return item
}

func (m *MyProjectList) getProjectDTO(orgID string, queryStr string) (*apistructs.PagingProjectDTO, error) {
	orgIDInt, err := strconv.ParseUint(m.ctxBdl.Identity.OrgID, 10, 64)
	if err != nil {
		return nil, err
	}
	req := apistructs.ProjectListRequest{
		OrgID:    orgIDInt,
		Query:    queryStr,
		PageNo:   m.State.PageNo,
		PageSize: m.State.PageSize,
		OrderBy:  "name",
		Asc:      true,
	}
	projectDTO, err := m.ctxBdl.Bdl.ListMyProject(m.ctxBdl.Identity.UserID, req)
	if err != nil {
		return nil, err
	}
	return projectDTO, nil
}

func (this *MyProjectList) getProjectsNum(orgID string, queryStr string) (int, error) {
	orgIntId, err := strconv.Atoi(orgID)
	if err != nil {
		return 0, err
	}
	req := apistructs.ProjectListRequest{
		OrgID:    uint64(orgIntId),
		PageNo:   1,
		PageSize: 1,
		Query:    queryStr,
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

func (m *MyProjectList) addDataList(datas *apistructs.PagingProjectDTO) error {
	var orgName string
	dataList := make([]ProItem, 0)
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

func (this *MyProjectList) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := this.GenComponentState(c); err != nil {
		return err
	}
	if this.ctxBdl.Identity.OrgID == "" {
		this.Props.Visible = false
		return nil
	}
	if this.State.ProNums != 0 {
		this.Props.Visible = true
	}
	this.Props.UseLoadMore = true
	this.Props.AlignCenter = true
	this.Props.Size = "small"
	this.Props.NoBorder = true
	this.Operations = map[string]interface{}{
		"changePageNo": map[string]interface{}{
			"key":      "changePageNo",
			"reload":   true,
			"fillMeta": "pageNo",
		},
	}

	queryStr := ""
	_, ok := this.State.Values["title"]
	if ok {
		queryStr = this.State.Values["title"].(string)
	}

	switch event.Operation {
	case apistructs.InitializeOperation:
		prosNum, err := this.getProjectsNum(this.ctxBdl.Identity.OrgID, queryStr)
		if err != nil {
			return err
		}
		if prosNum == 0 && queryStr == "" {
			this.Props.Visible = false
			return nil
		}
		this.State.PageNo = DefaultPageNo
		this.State.PageSize = DefaultPageSize
		projectDTO, err := this.getProjectDTO(this.ctxBdl.Identity.OrgID, queryStr)
		if err != nil {
			return err
		}
		this.State.Total = 0
		if projectDTO != nil {
			if err := this.addDataList(projectDTO); err != nil {
				return err
			}
			this.State.Total = projectDTO.Total
		}
	case apistructs.RenderingOperation:
		this.Data.List = make([]ProItem, 0)
		this.State.PageNo = DefaultPageNo
		this.State.PageSize = DefaultPageSize
		projectDTO, err := this.getProjectDTO(this.ctxBdl.Identity.OrgID, queryStr)
		if err != nil {
			return err
		}
		this.State.Total = 0
		if projectDTO != nil {
			if err := this.addDataList(projectDTO); err != nil {
				return err
			}
			this.State.Total = projectDTO.Total
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
		projectDTO, err := this.getProjectDTO(this.ctxBdl.Identity.OrgID, queryStr)
		if err != nil {
			return err
		}
		this.State.Total = 0
		if projectDTO != nil {
			if err := this.addDataList(projectDTO); err != nil {
				return err
			}
			this.State.Total = projectDTO.Total
		}
	}
	//sort.Sort(this.Data)
	return nil
}

func RenderCreator() protocol.CompRender {
	return &MyProjectList{}
}
