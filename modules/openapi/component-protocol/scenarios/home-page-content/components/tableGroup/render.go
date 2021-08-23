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

package tableGroup

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-content/i18n"
	"github.com/erda-project/erda/pkg/encoding/jsonparse"
)

const (
	DefaultPageNo    = 1
	DefaultPageSize  = 3
	DefaultIssueSize = 5
)

var issueTypeMap = map[string]string{
	"REQUIREMENT": "requirement",
	"TASK":        "task",
	"BUG":         "bug",
}

func (s ProData) Less(i, j int) bool {
	return s.List[i].Title.Props.DisplayName < s.List[j].Title.Props.DisplayName
}

func (s ProData) Swap(i, j int) {
	s.List[i], s.List[j] = s.List[j], s.List[i]
}

func (s ProData) Len() int {
	return len(s.List)
}

func (this *TableGroup) GenComponentState(c *apistructs.Component) error {
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

func (this *TableGroup) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (t *TableGroup) GetProsByPage() (*apistructs.PagingProjectDTO, error) {
	orgIDInt, err := strconv.ParseUint(t.ctxBdl.Identity.OrgID, 10, 64)
	if err != nil {
		return nil, err
	}
	req := apistructs.ProjectListRequest{
		OrgID:    uint64(orgIDInt),
		PageNo:   t.State.PageNo,
		PageSize: t.State.PageSize,
	}
	pageProDTO, err := t.ctxBdl.Bdl.ListDopProject(t.ctxBdl.Identity.UserID, req)
	if err != nil {
		return nil, err
	}
	return pageProDTO, nil
}

func (t *TableGroup) getOrgName() (name string, displayName string, err error) {
	var orgDTO *apistructs.OrgDTO
	orgDTO, err = t.ctxBdl.Bdl.GetOrg(t.ctxBdl.Identity.OrgID)
	if err != nil {
		return
	}
	if orgDTO == nil {
		return "", "", fmt.Errorf("failed to get org")
	}
	name = orgDTO.Name
	displayName = orgDTO.DisplayName
	return
}

func (t *TableGroup) getWorkbenchData() (*apistructs.WorkbenchResponse, error) {
	orgID, err := strconv.ParseUint(t.ctxBdl.Identity.OrgID, 10, 64)
	if err != nil {
		return nil, err
	}
	req := apistructs.WorkbenchRequest{
		OrgID:     orgID,
		PageSize:  t.State.PageSize,
		PageNo:    t.State.PageNo,
		IssueSize: 5,
	}
	res, err := t.ctxBdl.Bdl.GetWorkbenchData(t.ctxBdl.Identity.UserID, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *TableGroup) getProjectsNum(orgID string) (int, error) {
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

func (t *TableGroup) addWorkbenchData(datas *apistructs.WorkbenchResponse, orgName string, orgDisplayName string) {
	dataList := make([]ProItem, 0)
	i18nLocale := t.ctxBdl.Bdl.GetLocale(t.ctxBdl.Locale)
	for _, v := range datas.Data.List {
		pro := ProItem{}
		image := "frontImg_default_project_icon"
		if v.ProjectDTO.Logo != "" {
			image = v.ProjectDTO.Logo
		}
		pro.Title.Props = TitleProps{
			DisplayName: fmt.Sprintf("%s/%s", orgDisplayName, v.ProjectDTO.DisplayName),
			RenderType:  "linkText",
			Value: map[string]interface{}{
				"text": []interface{}{map[string]interface{}{"image": image}, map[string]interface{}{
					"text":         fmt.Sprintf("%s/%s", orgDisplayName, v.ProjectDTO.DisplayName),
					"operationKey": "toSpecificProject",
					"styleConfig": map[string]interface{}{
						"bold":     true,
						"color":    "black",
						"fontSize": "16px",
					},
				}},
				"isPureText": false,
			},
		}
		pro.Title.Operations = map[string]interface{}{
			"toSpecificProject": map[string]interface{}{
				"command": map[string]interface{}{
					"key":     "goto",
					"target":  "projectAllIssue",
					"jumpOut": true,
					"state": map[string]interface{}{
						"query": map[string]interface{}{
							"issueViewGroup__urlQuery": "eyJ2YWx1ZSI6ImthbmJhbiIsImNoaWxkcmVuVmFsdWUiOnsia2FuYmFuIjoiZGVhZGxpbmUifX0=",
						},
						"params": map[string]string{
							"projectId": strconv.FormatInt(int64(v.ProjectDTO.ID), 10),
							"orgName":   orgName,
						},
					},
				},
				"key":    "clieck",
				"reload": false,
				"show":   false,
			},
		}
		//pro.Title.IsPureTitle = false
		//if v.ProjectDTO.Logo != "" {
		//	pro.Title.PrefixImage = fmt.Sprintf("https:%s", v.ProjectDTO.Logo)
		//}
		//pro.Title.PrefixImage = ""
		//pro.Title.Title = fmt.Sprintf("%s/%s", orgName, v.ProjectDTO.DisplayName)
		//pro.Title.Level = 2
		pro.SubTitle.Title = i18nLocale.Get(i18n.I18nKeyProUnDoneIssue)
		pro.SubTitle.Level = 3
		pro.SubTitle.Size = "small"
		pro.Description.RenderType = "linkText"
		pro.Description.Visible = true
		pro.Description.Value = map[string]interface{}{
			"text": []interface{}{
				i18nLocale.Get(i18n.I18nKeyNowHave), map[string]interface{}{
					"text":        fmt.Sprintf(" %d ", v.TotalIssueNum),
					"styleConfig": map[string]interface{}{"bold": true, "fontSize": "16px"},
				}, i18nLocale.Get(i18n.I18nKeyIssueNoSpecified), map[string]interface{}{
					"text":        fmt.Sprintf(" %d ", v.UnSpecialIssueNum),
					"styleConfig": map[string]interface{}{"bold": true, "fontSize": "16px"},
				}, i18nLocale.Get(i18n.I18nKeyIssueExpired), map[string]interface{}{
					"text":        fmt.Sprintf(" %d ", v.ExpiredIssueNum),
					"styleConfig": map[string]interface{}{"bold": true, "fontSize": "16px"},
				}, i18nLocale.Get(i18n.I18nKeyIssueTodayExpired), map[string]interface{}{
					"text":        fmt.Sprintf(" %d ", v.ExpiredOneDayNum),
					"styleConfig": map[string]interface{}{"bold": true, "fontSize": "16px"},
				}, i18nLocale.Get(i18n.I18nKeyIssueTomorrowExpired), map[string]interface{}{
					"text":        fmt.Sprintf(" %d ", v.ExpiredTomorrowNum),
					"styleConfig": map[string]interface{}{"bold": true, "fontSize": "16px"},
				}, i18nLocale.Get(i18n.I18nKeyIssueSevenDayExpired), map[string]interface{}{
					"text":        fmt.Sprintf(" %d ", v.ExpiredSevenDayNum),
					"styleConfig": map[string]interface{}{"bold": true, "fontSize": "16px"},
				}, i18nLocale.Get(i18n.I18nKeyIssueThirtyDayExpired), map[string]interface{}{
					"text":        fmt.Sprintf(" %d ", v.ExpiredThirtyDayNum),
					"styleConfig": map[string]interface{}{"bold": true, "fontSize": "16px"},
				}, i18nLocale.Get(i18n.I18nKeyIssueFeatureExpired), map[string]interface{}{
					"text":        fmt.Sprintf(" %d ", v.FeatureDayNum),
					"styleConfig": map[string]interface{}{"bold": true, "fontSize": "16px"},
				},
			},
		}
		pro.Description.TextStyleName = map[string]interface{}{
			"color-text-light-desc": true,
		}
		//pro.Description.Value = fmt.Sprintf("当前您还有 %d 个事项待完成，其中 已过期: %d，本日到期: %d，7日内到期: %d，30日内到期: %d",
		//	v.TotalIssueNum, v.ExpiredIssueNum, v.ExpiredOneDayNum, v.ExpiredSevenDayNum, v.ExpiredThirtyDayNum)
		pro.Table.Props = map[string]interface{}{
			"rowKey": v.ProjectDTO.ID,
			"columns": []interface{}{
				map[string]interface{}{"title": "", "dataIndex": "name", "width": 600},
				map[string]interface{}{"title": "", "dataIndex": "planFinishedAt"},
			},
			"showHeader": false,
			"pagination": false,
			"styleNames": map[string]bool{
				"no-border": true,
			},
			"size": "small",
		}
		issueDatas := make([]IssueItem, 0)
		for _, issue := range v.IssueList {
			issueItem := IssueItem{
				Id:        issue.ID,
				ProjectId: v.ProjectDTO.ID,
				Type:      issueTypeMap[issue.Type.String()],
				Name: IssueName{
					RenderType:  "textWithIcon",
					PrefixIcon:  fmt.Sprintf("ISSUE_ICON.issue.%s", issue.Type.String()),
					Value:       issue.Title,
					HoverActive: "hover-active",
				},
				OrgName: orgName,
			}
			issueItem.PlanFinishedAt = i18nLocale.Get(i18n.I18nKeyIssueNotSpecifiedDate)
			if issue.PlanFinishedAt != nil {
				issueItem.PlanFinishedAt = issue.PlanFinishedAt.Format("2006-01-02")
			}
			issueDatas = append(issueDatas, issueItem)
		}
		pro.Table.Data.List = issueDatas
		click := ClickOperation{
			Key:    "clickRow",
			Reload: false,
		}
		click.Command.Key = "goto"
		click.Command.Target = "projectIssueDetail"
		click.Command.JumpOut = true
		pro.Table.Operations = map[string]interface{}{
			"clickRow": click,
		}
		leftIssueNum := v.TotalIssueNum - len(v.IssueList)
		projectOperation := ToSpecificProjectOperation{
			Key:    "click",
			Reload: false,
			Show:   false,
		}
		projectOperation.Command.Key = "goto"
		projectOperation.Command.Target = "projectAllIssue"
		projectOperation.Command.JumpOut = true
		projectOperation.Command.State.Query.IssueViewGroupUrlQuery = "eyJ2YWx1ZSI6ImthbmJhbiIsImNoaWxkcmVuVmFsdWUiOnsia2FuYmFuIjoiZGVhZGxpbmUifX0="
		projectOperation.Command.State.Query.IssueTableUrlQuery = "eyJwYWdlTm8iOjF9"
		projectOperation.Command.State.Query.IssueFilterUrlQuery = t.generateIssueUrlQuery()
		projectOperation.Command.State.Params.ProjectId = strconv.FormatInt(int64(v.ProjectDTO.ID), 10)
		projectOperation.Command.State.Params.OrgName = orgName
		projectOperation.Command.Visible = true
		pro.ExtraInfo = ExtraInfo{
			Props: ExtraProps{
				RenderType: "linkText",
				Value: Value{
					Text: []ValueText{
						{Text: fmt.Sprintf("%s%d%s ", i18nLocale.Get(i18n.I18nKeyViewRemaining), leftIssueNum, i18nLocale.Get(i18n.I18nKeySomeEvent)), OperationKey: "toSpecificProject", Icon: "double-right"},
					},
				},
			},
			Operations: map[string]interface{}{
				"toSpecificProject": projectOperation,
			},
		}
		dataList = append(dataList, pro)
	}
	t.Data.List = dataList
}

func (t *TableGroup) generateIssueUrlQuery() string {
	queryMap := map[string]interface{}{
		"stateBelongs": []apistructs.IssueStateBelong{apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking,
			apistructs.IssueStateBelongWontfix, apistructs.IssueStateBelongReopen},
		"assigneeIDs": []string{t.ctxBdl.Identity.UserID},
	}
	queryMapStr := jsonparse.JsonOneLine(queryMap)
	return base64.StdEncoding.EncodeToString([]byte(queryMapStr))
}

func (t *TableGroup) setBaseComponentValue() {
	t.Type = "TableGroup"
	t.Operations = map[string]interface{}{
		"changePageNo": ChangePageNoOperation{
			Key:      "changePageNo",
			Reload:   true,
			FillMeta: "pageNo",
		},
	}
}

func (t *TableGroup) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := t.GenComponentState(c); err != nil {
		return err
	}
	if err := t.SetCtxBundle(ctx); err != nil {
		return err
	}

	t.setBaseComponentValue()
	if t.ctxBdl.Identity.OrgID == "" {
		t.Props.Visible = false
		return nil
	}
	t.Props.Visible = true
	t.Props.IsLoadMore = true
	orgName, orgDisplayName, err := t.getOrgName()
	if err != nil {
		return err
	}
	switch event.Operation {
	case apistructs.InitializeOperation:
		t.State.PageNo = DefaultPageNo
		t.State.PageSize = DefaultPageSize
		workDatas, err := t.getWorkbenchData()
		if err != nil {
			return err
		}
		t.State.Total = 0
		t.Data.List = make([]ProItem, 0)
		if workDatas != nil {
			t.State.Total = workDatas.Data.TotalProject
			t.addWorkbenchData(workDatas, orgName, orgDisplayName)
		}
	case apistructs.ChangeOrgsPageNoOperationKey:
		if t.State.PageNo <= 0 || t.State.PageSize <= 0 {
			return fmt.Errorf("invalid page size or no")
		}
		//if t.State.Total <= (t.State.PageSize) * (t.State.PageNo-1) {
		//	return fmt.Errorf("查询数超过项目总数")
		//}
		var pageData OperationData
		dataBody, err := json.Marshal(event.OperationData)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(dataBody, &pageData); err != nil {
			return err
		}
		t.State.PageNo = pageData.Meta.PageNo.PageNo
		workDatas, err := t.getWorkbenchData()
		if err != nil {
			return err
		}
		t.State.Total = 0
		t.Data.List = make([]ProItem, 0)
		if workDatas != nil {
			t.State.Total = workDatas.Data.TotalProject
			t.addWorkbenchData(workDatas, orgName, orgDisplayName)
		}
	}
	projectNum, err := t.getProjectsNum(t.ctxBdl.Identity.OrgID)
	if err != nil {
		return err
	}
	t.State.ProsNum = projectNum
	if t.State.Total == 0 {
		t.Props.Visible = false
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &TableGroup{}
}
