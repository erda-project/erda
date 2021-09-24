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

package issueTable

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/bdl"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-manage/issueViewGroup"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/pkg/strutil"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/hooks/posthandle"
)

type Progress struct {
	Value      string `json:"value"`
	RenderType string `json:"renderType"`
}
type Severity struct {
	Value       string                 `json:"value"`
	RenderType  string                 `json:"renderType"`
	PrefixIcon  string                 `json:"prefixIcon"`
	Operations  map[string]interface{} `json:"operations"`
	Disabled    bool                   `json:"disabled"`
	DisabledTip string                 `json:"disabledTip"`
}

type Title TableColumnMultiple

type TableColumnMultiple struct {
	RenderType string          `json:"renderType,omitempty"`
	Renders    [][]interface{} `json:"renders,omitempty"`
}

type TableColumnTextWithIcon struct {
	RenderType string `json:"renderType,omitempty"`
	Value      string `json:"value,omitempty"`
	PrefixIcon string `json:"prefixIcon,omitempty"`
}

type TableColumnTagsRow struct {
	RenderType string                  `json:"renderType,omitempty"`
	Value      []TableColumnTagsRowTag `json:"value,omitempty"`
	ShowCount  int                     `json:"showCount,omitempty"`
}

type TableColumnTagsRowTag struct {
	Color string `json:"color,omitempty"`
	Label string `json:"label,omitempty"`
}

type State struct {
	Operations  map[string]interface{} `json:"operations"`
	PrefixIcon  string                 `json:"prefixIcon"`
	Value       string                 `json:"value"`
	RenderType  string                 `json:"renderType"`
	Disabled    bool                   `json:"disabled"`
	DisabledTip string                 `json:"disabledTip"`
}

type Priority struct {
	Operations  map[string]interface{} `json:"operations"`
	PrefixIcon  string                 `json:"prefixIcon"`
	Value       string                 `json:"value"`
	RenderType  string                 `json:"renderType"`
	Disabled    bool                   `json:"disabled"`
	DisabledTip string                 `json:"disabledTip"`
}

type Deadline struct {
	RenderType     string                 `json:"renderType"`
	Value          string                 `json:"value"`
	NoBorder       bool                   `json:"noBorder"`
	DisabledBefore string                 `json:"disabledBefore"`
	DisabledAfter  string                 `json:"disabledAfter"`
	Operations     map[string]interface{} `json:"operations"`
}

type Assignee struct {
	Value       string                 `json:"value"`
	RenderType  string                 `json:"renderType"`
	Scope       string                 `json:"scope"`
	Operations  map[string]interface{} `json:"operations"`
	Disabled    bool                   `json:"disabled"`
	DisabledTip string                 `json:"disabledTip"`
}
type TableItem struct {
	//Assignee    map[string]string `json:"assignee"`
	Id          string   `json:"id"`
	IterationID int64    `json:"iterationID"`
	Priority    Priority `json:"priority"`
	Progress    Progress `json:"progress,omitempty"`
	Severity    Severity `json:"severity,omitempty"`
	Complexity  string   `json:"complexity,omitempty"`
	State       State    `json:"state"`
	Title       Title    `json:"title"`
	Type        string   `json:"type"`
	Deadline    Deadline `json:"deadline"`
	Assignee    Assignee `json:"assignee"`
	ClosedAt    ClosedAt `json:"closedAt"`
}

type ClosedAt struct {
	RenderType string `json:"renderType"`
	Value      string `json:"value"`
	NoBorder   bool   `json:"noBorder"`
}

type PriorityOperationData struct {
	Meta struct {
		Priority string `json:"priority"`
		ID       string `json:"id"`
	} `json:"meta"`
}
type DeadlineOperationData struct {
	Meta struct {
		DeadlineValue string `json:"deadlineValue"`
		ID            string `json:"id"`
	} `json:"meta"`
}
type StateOperationData struct {
	Meta struct {
		State string `json:"state"`
		ID    string `json:"id"`
	} `json:"meta"`
}
type SeverityOperationData struct {
	Meta struct {
		Severity string `json:"severity"`
		ID       string `json:"id"`
	} `json:"meta"`
}
type AssigneeOperationData struct {
	Meta struct {
		Assignee string `json:"assignee"`
		ID       string `json:"id"`
	} `json:"meta"`
}

type ComponentAction struct {
	base.DefaultProvider
	labels  []apistructs.ProjectLabel
	isGuest bool
	userMap map[string]apistructs.UserInfo
}

var (
	priorityIcon = map[apistructs.IssuePriority]string{
		apistructs.IssuePriorityUrgent: "ISSUE_ICON.priority.URGENT", // 紧急
		apistructs.IssuePriorityHigh:   "ISSUE_ICON.priority.HIGH",   // 高
		apistructs.IssuePriorityNormal: "ISSUE_ICON.priority.NORMAL", // 中
		apistructs.IssuePriorityLow:    "ISSUE_ICON.priority.LOW",    // 低
	}

	stateIcon = map[string]string{
		string(apistructs.IssueStateOpen):     "ISSUE_ICON.state.OPEN",
		string(apistructs.IssueStateWorking):  "ISSUE_ICON.state.WORKING",
		string(apistructs.IssueStateTesting):  "ISSUE_ICON.state.TESTING",
		string(apistructs.IssueStateDone):     "ISSUE_ICON.state.DONE",
		string(apistructs.IssueStateResolved): "ISSUE_ICON.state.RESOLVED",
		string(apistructs.IssueStateReopen):   "ISSUE_ICON.state.REOPEN",
		string(apistructs.IssueStateWontfix):  "ISSUE_ICON.state.WONTFIX",
		string(apistructs.IssueStateDup):      "ISSUE_ICON.state.DUP",
		string(apistructs.IssueStateClosed):   "ISSUE_ICON.state.CLOSED",
	}
)

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	// visible
	visible := true
	if v, ok := c.State["issueViewGroupValue"]; ok {
		if viewType, ok := v.(string); ok {
			if viewType != issueViewGroup.ViewTypeTable {
				visible = false
				c.Props = map[string]interface{}{}
				c.Props.(map[string]interface{})["visible"] = visible
				return nil
			}
		}
	}

	sdk := cputil.SDK(ctx)

	isGuest, err := ca.CheckUserPermission(ctx)
	if err != nil {
		return err
	}
	ca.isGuest = isGuest

	projectid, err := strconv.ParseUint(sdk.InParams["projectId"].(string), 10, 64)
	orgid, err := strconv.ParseUint(sdk.Identity.OrgID, 10, 64)
	if strutil.HasPrefixes(event.Operation.String(), "changePriorityTo") {
		priority := apistructs.IssuePriorityLow
		switch event.Operation.String() {
		case "changePriorityToaURGENT":
			priority = apistructs.IssuePriorityUrgent
		case "changePriorityTobHIGH":
			priority = apistructs.IssuePriorityHigh
		case "changePriorityTocNORMAL":
			priority = apistructs.IssuePriorityNormal
		case "changePriorityTodLOW":
			priority = apistructs.IssuePriorityLow
		}
		od, _ := json.Marshal(event.OperationData)
		operationData := PriorityOperationData{}
		if err := json.Unmarshal(od, &operationData); err != nil {
			return err
		}
		id, err := strconv.ParseUint(operationData.Meta.ID, 10, 64)
		if err != nil {
			return err
		}
		is, err := bdl.Bdl.GetIssue(id)
		if err != nil {
			logrus.Errorf("get issue failed, id:%v, err:%v", id, err)
			return err
		}
		is.Priority = priority
		err = bdl.Bdl.UpdateIssueTicketUser(sdk.Identity.UserID, is.ConvertToIssueUpdateReq(), uint64(is.ID))
		if err != nil {
			logrus.Errorf("update issue failed, id:%v, err:%v", id, err)
			return err
		}
	}
	if event.Operation.String() == "changeDeadline" {
		od, _ := json.Marshal(event.OperationData)
		operationData := DeadlineOperationData{}
		if err := json.Unmarshal(od, &operationData); err != nil {
			return err
		}
		id, err := strconv.ParseUint(operationData.Meta.ID, 10, 64)
		if err != nil {
			return err
		}
		deadline, err := time.Parse(time.RFC3339, operationData.Meta.DeadlineValue)
		if err != nil {
			return err
		}
		is, err := bdl.Bdl.GetIssue(id)
		if err != nil {
			logrus.Errorf("get issue failed, id:%v, err:%v", id, err)
			return err
		}
		is.PlanFinishedAt = &deadline
		err = bdl.Bdl.UpdateIssueTicketUser(sdk.Identity.UserID, is.ConvertToIssueUpdateReq(), uint64(is.ID))
		if err != nil {
			logrus.Errorf("update issue failed, id:%v, err:%v", id, err)
			return err
		}
	}

	if strutil.HasPrefixes(event.Operation.String(), "changeStateTo") {
		od, _ := json.Marshal(event.OperationData)
		operationData := StateOperationData{}
		if err := json.Unmarshal(od, &operationData); err != nil {
			return err
		}
		id, err := strconv.ParseUint(operationData.Meta.ID, 10, 64)
		if err != nil {
			return err
		}
		stateid, err := strconv.ParseInt(operationData.Meta.State, 10, 64)
		if err != nil {
			return err
		}
		is, err := bdl.Bdl.GetIssue(id)
		if err != nil {
			logrus.Errorf("get issue failed, id:%v, err:%v", id, err)
			return err
		}
		is.State = stateid
		err = bdl.Bdl.UpdateIssueTicketUser(sdk.Identity.UserID, is.ConvertToIssueUpdateReq(), uint64(is.ID))
		if err != nil {
			logrus.Errorf("update issue failed, id:%v, err:%v", id, err)
			return err
		}
	}
	if strutil.HasPrefixes(event.Operation.String(), "changeSeverityTo") {
		od, _ := json.Marshal(event.OperationData)
		operationData := SeverityOperationData{}
		if err := json.Unmarshal(od, &operationData); err != nil {
			return err
		}
		id, err := strconv.ParseUint(operationData.Meta.ID, 10, 64)
		if err != nil {
			return err
		}
		severity := apistructs.IssueSeverity(operationData.Meta.Severity)
		is, err := bdl.Bdl.GetIssue(id)
		if err != nil {
			logrus.Errorf("get issue failed, id:%v, err:%v", id, err)
			return err
		}
		is.Severity = severity
		err = bdl.Bdl.UpdateIssueTicketUser(sdk.Identity.UserID, is.ConvertToIssueUpdateReq(), uint64(is.ID))
		if err != nil {
			logrus.Errorf("update issue failed, id:%v, err:%v", id, err)
			return err
		}
	}
	if strutil.HasPrefixes(event.Operation.String(), "updateAssignee") {
		od, _ := json.Marshal(event.OperationData)
		operationData := AssigneeOperationData{}
		if err := json.Unmarshal(od, &operationData); err != nil {
			return err
		}
		id, err := strconv.ParseUint(operationData.Meta.ID, 10, 64)
		if err != nil {
			return err
		}
		assgignee := operationData.Meta.Assignee
		is, err := bdl.Bdl.GetIssue(id)
		if err != nil {
			logrus.Errorf("get issue failed, id:%v, err:%v", id, err)
			return err
		}
		is.Assignee = assgignee
		err = bdl.Bdl.UpdateIssueTicketUser(sdk.Identity.UserID, is.ConvertToIssueUpdateReq(), uint64(is.ID))
		if err != nil {
			logrus.Errorf("update issue failed, id:%v, err:%v", id, err)
			return err
		}
	}
	userids := []string{}
	cond := apistructs.IssuePagingRequest{}
	filterCond, ok := c.State["filterConditions"]
	if ok {
		filterCondS, err := json.Marshal(filterCond)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(filterCondS, &cond); err != nil {
			return err
		}
		resetPageInfo(&cond, c.State)
	} else {
		issuetype := sdk.InParams["fixedIssueType"].(string)
		switch issuetype {
		case string(apistructs.IssueTypeRequirement):
			cond.Type = []apistructs.IssueType{apistructs.IssueTypeRequirement}
		case string(apistructs.IssueTypeTask):
			cond.Type = []apistructs.IssueType{apistructs.IssueTypeTask}
		case string(apistructs.IssueTypeBug):
			cond.Type = []apistructs.IssueType{apistructs.IssueTypeBug}
		default:
			cond.Type = []apistructs.IssueType{apistructs.IssueTypeRequirement, apistructs.IssueTypeTask, apistructs.IssueTypeBug, apistructs.IssueTypeEpic}
		}
		cond.OrgID = int64(orgid)
		cond.PageNo = 1
		if c.State != nil {
			if _, ok := c.State["pageNo"]; ok {
				cond.PageNo = uint64(c.State["pageNo"].(float64))
			}
		}
		cond.PageSize = 10
		cond.ProjectID = projectid
		cond.IssueListRequest.IdentityInfo.UserID = sdk.Identity.UserID
	}
	if event.Operation.String() == "changePageNo" {
		cond.PageNo = 1
		if c.State != nil {
			if _, ok := c.State["pageNo"]; ok {
				cond.PageNo = uint64(c.State["pageNo"].(float64))
			}
		}
		cond.PageSize = uint64(c.State["pageSize"].(float64))
	} else if event.Operation == cptype.OperationKey(apistructs.InitializeOperation) {
		cond.PageNo = 1
		if urlquery, ok := sdk.InParams["issueTable__urlQuery"]; ok {
			urlquery_d, err := base64.StdEncoding.DecodeString(urlquery.(string))
			if err != nil {
				return err
			}
			querymap := map[string]interface{}{}
			if err := json.Unmarshal(urlquery_d, &querymap); err != nil {
				return err
			}
			if pageNoInQuery, ok := querymap["pageNo"]; ok {
				cond.PageNo = uint64(pageNoInQuery.(float64))
			}
			if pageSize, ok := querymap["pageSize"]; ok {
				cond.PageSize = uint64(pageSize.(float64))
			}
		}
	} else if event.Operation == cptype.OperationKey(apistructs.RenderingOperation) {
		cond.PageNo = 1
	}

	// check reset pageNo
	if c.State != nil && resetPageNoByFilterCondition(event.Operation.String(), TableItem{}, c.State) {
		if _, ok := c.State["pageNo"]; ok {
			cond.PageNo = uint64(c.State["pageNo"].(float64))
		}
	}

	var (
		pageTotal uint64
		r         *apistructs.IssuePagingResponse
	)
	r, err = bdl.Bdl.PageIssues(cond)
	if err != nil {
		return err
	}
	// if pageTotal < cond.PageNo, to reset the cond.PageNo = 1,
	// and return the first page of data
	pageTotal = getTotalPage(r.Data.Total, cond.PageSize)
	if pageTotal < cond.PageNo {
		cond.PageNo = 1
		r, err = bdl.Bdl.PageIssues(cond)
		if err != nil {
			return err
		}
	}

	for _, p := range r.Data.List {
		userids = append(userids, p.Assignee)
	}
	// 获取全部用户
	userids = strutil.DedupSlice(userids, true)
	uInfo, err := posthandle.GetUsers(userids, false)
	if err != nil {
		return err
	}
	ca.userMap = make(map[string]apistructs.UserInfo)
	for _, v := range uInfo {
		ca.userMap[v.ID] = v
	}

	labels, err := bdl.Bdl.Labels("issue", projectid, sdk.Identity.UserID)
	if err != nil {
		return err
	}
	ca.labels = labels.List
	var l []TableItem
	for _, data := range r.Data.List {
		l = append(l, *ca.buildTableItem(ctx, &data))
	}
	c.Data = map[string]interface{}{}
	c.Data["list"] = l
	progressCol := ""
	if len(cond.Type) == 1 && cond.Type[0] == apistructs.IssueTypeRequirement {
		progressCol = `{
            "width": 100,
            "dataIndex": "progress",
            "title": "进度"
        },`
	}

	severityCol, closedAtCol := "", ""
	if len(cond.Type) == 1 && cond.Type[0] == apistructs.IssueTypeBug {
		severityCol = `{ "title": "严重程度", "dataIndex": "severity", "width": 100 },`
		closedAtCol = `,{ "title": "关闭日期", "dataIndex": "closedAt", "width": 100 }`
	}
	props := `{
    "columns": [
		{
			"dataIndex": "id",
			"title": "ID",
			"width": 90
        },
        {
            "dataIndex": "title",
            "title": "` + cputil.I18n(ctx, "title") + `"
        },` +
		progressCol +
		severityCol +
		`{
            "width": 100,
            "dataIndex": "complexity",
            "title": "` + cputil.I18n(ctx, "complexity") + `"
        },
        {
            "width": 100,
            "dataIndex": "priority",
            "title": "` + cputil.I18n(ctx, "priority") + `"
        },
        {
            "width": 110,
            "dataIndex": "state",
            "title": "` + cputil.I18n(ctx, "state") + `"
        },
        {
            "width": 120,
            "dataIndex": "assignee",
            "title": "` + cputil.I18n(ctx, "assignee") + `"
        },
        {
            "width": 100,
            "dataIndex": "deadline",
            "title": "` + cputil.I18n(ctx, "deadline") + `"
        }` +
		closedAtCol +
		`],
    "rowKey": "id",
	"scroll": {"x": 1200},
	"pageSizeOptions": ["10", "20", "50", "100"]
}`
	var propsI interface{}
	if err := json.Unmarshal([]byte(props), &propsI); err != nil {
		return err
	}
	c.Props = propsI
	c.Operations = map[string]interface{}{
		"changePageNo": map[string]interface{}{
			"key":    "changePageNo",
			"reload": true,
		},
		"changePageSize": map[string]interface{}{
			"key":    "changePageSize",
			"reload": true,
		},
	}
	c.Props.(map[string]interface{})["visible"] = visible
	(*gs)[protocol.GlobalInnerKeyUserIDs.String()] = userids
	if c.State == nil {
		c.State = map[string]interface{}{}
	}
	c.State["total"] = r.Data.Total
	c.State["pageNo"] = cond.PageNo
	c.State["pageSize"] = cond.PageSize
	urlquery := fmt.Sprintf(`{"pageNo":%d, "pageSize":%d}`, cond.PageNo, cond.PageSize)
	c.State["issueTable__urlQuery"] = base64.StdEncoding.EncodeToString([]byte(urlquery))
	return nil
}

func (ca *ComponentAction) buildTableItem(ctx context.Context, data *apistructs.Issue) *TableItem {
	var issuestate *apistructs.IssueStateButton
	for _, s := range data.IssueButton {
		if s.StateID == data.State {
			issuestate = &s
			break
		}
	}
	titleColumn := ca.getTitleColumn(data)
	progress := Progress{
		RenderType: "progress",
		Value:      "",
	}
	if data.IssueSummary != nil && data.IssueSummary.DoneCount+data.IssueSummary.ProcessingCount > 0 {
		progress.Value = fmt.Sprintf("%d", int(100*(float64(data.IssueSummary.DoneCount)/float64(data.IssueSummary.DoneCount+data.IssueSummary.ProcessingCount))))
	}

	severityOps := map[string]interface{}{}
	severityAuto := map[apistructs.IssueSeverity]string{
		apistructs.IssueSeverityFatal:   "a",
		apistructs.IssueSeveritySerious: "b",
		apistructs.IssueSeverityNormal:  "c",
		apistructs.IssueSeveritySlight:  "d",
		apistructs.IssueSeverityLow:     "e",
	}
	for _, s := range []apistructs.IssueSeverity{

		apistructs.IssueSeverityFatal,
		apistructs.IssueSeveritySerious,
		apistructs.IssueSeverityNormal,
		apistructs.IssueSeveritySlight,
		apistructs.IssueSeverityLow} {
		severityOps["changeSeverityTo"+severityAuto[s]+string(s)] = map[string]interface{}{
			"key":        "changeSeverityTo" + severityAuto[s] + string(s),
			"reload":     true,
			"text":       cputil.I18n(ctx, string(s)),
			"prefixIcon": "ISSUE_ICON.severity." + string(s),
			"meta": map[string]string{
				"id":       strconv.FormatInt(data.ID, 10),
				"severity": string(s),
			},
		}
	}
	severity := Severity{
		RenderType: "operationsDropdownMenu",
		Value:      cputil.I18n(ctx, strings.ToLower(string(data.Severity))),
		PrefixIcon: "ISSUE_ICON.severity." + string(data.Severity),
		Operations: severityOps,
		Disabled:   ca.isGuest,
		DisabledTip: map[bool]string{
			true: "无权限",
		}[ca.isGuest],
	}
	stateOperations := map[string]interface{}{}
	stateAllDisable := true
	for i, s := range data.IssueButton {
		if s.Permission {
			stateAllDisable = false
		}
		if ca.isGuest {
			stateAllDisable = true
		}
		if s.Permission {
			stateOperations["changeStateTo"+strconv.Itoa(i)+s.StateName] = map[string]interface{}{
				"meta": map[string]string{
					"state": strconv.FormatInt(s.StateID, 10),
					"id":    strconv.FormatInt(data.ID, 10),
				},
				"prefixIcon": stateIcon[string(s.StateBelong)],
				"text":       s.StateName,
				"reload":     true,
				"key":        "changeStateTo" + strconv.Itoa(i) + s.StateName,
				"disabled":   !s.Permission,
				"disabledTip": map[bool]string{
					false: "无法转移",
				}[s.Permission],
			}
		}
	}
	AssigneeMapOperations := map[string]interface{}{}
	AssigneeMapOperations["onChange"] = map[string]interface{}{
		"meta": map[string]string{
			"assignee": "",
			"id":       strconv.FormatInt(data.ID, 10),
		},
		"text":     ca.userMap[data.Assignee].Nick,
		"reload":   true,
		"key":      "updateAssignee",
		"disabled": false,
		"fillMeta": "assignee",
	}
	deadline := Deadline{
		RenderType: "datePicker",
		Value:      "",
		NoBorder:   true,
		Operations: map[string]interface{}{
			"onChange": map[string]interface{}{
				"key":      "changeDeadline",
				"reload":   true,
				"disabled": false,
				"fillMeta": "deadlineValue",
				"meta": map[string]string{
					"id":            strconv.FormatInt(data.ID, 10),
					"deadlineValue": "",
				},
			},
		},
	}
	if data.PlanFinishedAt != nil {
		deadline.Value = data.PlanFinishedAt.Format(time.RFC3339)
	}
	if data.PlanStartedAt != nil {
		deadline.DisabledBefore = data.PlanStartedAt.Format(time.RFC3339)
	}
	closedAt := ClosedAt{
		RenderType: "datePicker",
		Value:      "",
		NoBorder:   true,
	}
	if data.FinishTime != nil {
		closedAt.Value = data.FinishTime.Format(time.RFC3339)
	}
	state := State{
		Operations: stateOperations,
		RenderType: "operationsDropdownMenu",
		Disabled:   stateAllDisable,
		DisabledTip: map[bool]string{
			true: "无权限",
		}[stateAllDisable],
	}
	if issuestate != nil {
		state.PrefixIcon = stateIcon[string(issuestate.StateBelong)]
		state.Value = issuestate.StateName
	}
	return &TableItem{
		//Assignee:    map[string]string{"value": data.Assignee, "renderType": "userAvatar"},
		Id:          strconv.FormatInt(data.ID, 10),
		IterationID: data.IterationID,
		Type:        string(data.Type),
		Progress:    progress,
		Severity:    severity,
		Complexity:  cputil.I18n(ctx, string(data.Complexity)),
		Priority: Priority{
			Value:      cputil.I18n(ctx, strings.ToLower(string(data.Priority))),
			RenderType: "operationsDropdownMenu",
			PrefixIcon: priorityIcon[data.Priority],
			Operations: map[string]interface{}{
				"changePriorityTodLOW": map[string]interface{}{
					"meta": map[string]string{
						"id":       strconv.FormatInt(data.ID, 10),
						"priority": "LOW",
					},
					"prefixIcon": priorityIcon[apistructs.IssuePriorityLow],
					"text":       cputil.I18n(ctx, "low"),
					"reload":     true,
					"key":        "changePriorityTodLOW",
				}, "changePriorityTocNORMAL": map[string]interface{}{
					"meta": map[string]string{
						"id":       strconv.FormatInt(data.ID, 10),
						"priority": "NORMAL",
					},
					"prefixIcon": priorityIcon[apistructs.IssuePriorityNormal],
					"text":       cputil.I18n(ctx, "normal"),
					"reload":     true,
					"key":        "changePriorityTocNORMAL",
				}, "changePriorityTobHIGH": map[string]interface{}{
					"meta": map[string]string{
						"id":       strconv.FormatInt(data.ID, 10),
						"priority": "HIGH",
					},
					"prefixIcon": priorityIcon[apistructs.IssuePriorityHigh],
					"text":       cputil.I18n(ctx, "high"),
					"reload":     true,
					"key":        "changePriorityTobHIGH",
				},
				"changePriorityToaURGENT": map[string]interface{}{
					"meta": map[string]string{
						"id":       strconv.FormatInt(data.ID, 10),
						"priority": "URGENT",
					},
					"prefixIcon": priorityIcon[apistructs.IssuePriorityUrgent],
					"text":       cputil.I18n(ctx, "urgent"),
					"reload":     true,
					"key":        "changePriorityToaURGENT",
				},
			},
			Disabled: ca.isGuest,
			DisabledTip: map[bool]string{
				true: "无权限",
			}[ca.isGuest],
		},
		State: state,
		Assignee: Assignee{
			Value:      data.Assignee,
			RenderType: "memberSelector",
			Scope:      "project",
			Operations: AssigneeMapOperations,
			Disabled:   ca.isGuest,
			DisabledTip: map[bool]string{
				true: "无权限",
			}[ca.isGuest],
		},
		Title:    titleColumn,
		Deadline: deadline,
		ClosedAt: closedAt,
	}
}

func (ca *ComponentAction) getTitleColumn(issue *apistructs.Issue) Title {
	var tags []TableColumnTagsRowTag
	for _, label := range issue.Labels {
		for _, labelDef := range ca.labels {
			if label == labelDef.Name {
				tags = append(tags, TableColumnTagsRowTag{Color: labelDef.Color, Label: labelDef.Name})
			}
		}
	}
	return Title{
		RenderType: "multiple",
		Renders: [][]interface{}{
			{
				TableColumnTextWithIcon{
					RenderType: "textWithIcon",
					Value:      issue.Title,
					PrefixIcon: getPrefixIcon(string(issue.Type)),
				},
			},
			{
				TableColumnTagsRow{
					RenderType: "tagsRow",
					Value:      tags,
					ShowCount:  5,
				},
			},
		},
	}
}

// GetUserPermission  check Guest permission
func (ca *ComponentAction) CheckUserPermission(ctx context.Context) (bool, error) {
	sdk := cputil.SDK(ctx)
	isGuest := false
	projectID := sdk.InParams["projectId"].(string)
	scopeRole, err := bdl.Bdl.ScopeRoleAccess(sdk.Identity.UserID, &apistructs.ScopeRoleAccessRequest{
		Scope: apistructs.Scope{
			Type: apistructs.ProjectScope,
			ID:   projectID,
		},
	})
	if err != nil {
		return false, err
	}
	for _, role := range scopeRole.Roles {
		if role == "Guest" {
			isGuest = true
		}
	}
	return isGuest, nil
}

func getPrefixIcon(_type string) string {
	return "ISSUE_ICON.issue." + _type
}

func resetPageNoByFilterCondition(event string, filter interface{}, state map[string]interface{}) bool {
	v := reflect.ValueOf(filter).Type()
	for i := 0; i < v.NumField(); i++ {
		if strutil.Contains(event, v.Field(i).Name) {
			if v, ok := state[v.Field(i).Name]; ok && v != nil {
				return false
			}
		}
	}
	return true
}

// getTotalPage get total page
func getTotalPage(total, pageSize uint64) (page uint64) {
	if pageSize == 0 {
		return 0
	}
	if total%pageSize == 0 {
		return total / pageSize
	}
	return total/pageSize + 1
}

func init() {
	base.InitProviderWithCreator("issue-manage", "issueTable", func() servicehub.Provider {
		return &ComponentAction{}
	})
}

func resetPageInfo(req *apistructs.IssuePagingRequest, state map[string]interface{}) {
	req.PageSize = 10
	if _, ok := state["pageNo"]; ok {
		req.PageNo = uint64(state["pageNo"].(float64))
	}
	if _, ok := state["pageSize"]; ok {
		req.PageSize = uint64(state["pageSize"].(float64))
	}
}
