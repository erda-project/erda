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
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/bdl"
	"github.com/erda-project/erda/internal/apps/dop/component-protocol/components/common"
	"github.com/erda-project/erda/internal/apps/dop/component-protocol/components/issue-manage/common/gshelper"
	"github.com/erda-project/erda/internal/apps/dop/component-protocol/standard-components/issueFilter"
	"github.com/erda-project/erda/internal/apps/dop/component-protocol/types"
	issuemodel "github.com/erda-project/erda/internal/apps/dop/providers/issue/core/common"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	protocol "github.com/erda-project/erda/internal/core/openapi/legacy/component-protocol"
	"github.com/erda-project/erda/pkg/strutil"
)

type Progress TableColumnMultiple

type TableItemForShow map[string]interface{}

func (item *TableItem) toColumnDataRow() map[string]interface{} {
	data := make(map[string]interface{})
	item.BatchOperations = []string{"delete"}
	cputil.MustObjJSONTransfer(item, &data)
	// delete self properties field
	delete(data, "properties")
	// treat all custom properties as TextBlock
	for _, property := range item.Properties {
		block := TextBlock{
			Value:      (*issuemodel.PropertyInstanceForShow)(property).String(),
			RenderType: "text",
		}
		if property.PropertyType == pb.PropertyTypeEnum_Person && block.Value != "" {
			block.RenderType = "userAvatar"
		}
		data[makePropertyColumnName(property)] = block
	}
	return data
}

func makePropertyColumnName(property *pb.IssuePropertyExtraProperty) string {
	return "property_" + property.PropertyName
}

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	sdk := cputil.SDK(ctx)
	issueQuery := ctx.Value(types.IssueQuery).(query.Interface)
	issueSvc := ctx.Value(types.IssueService).(pb.IssueCoreServiceServer)
	identity := ctx.Value(types.IdentitiyService).(userpb.UserServiceServer)
	ca.issueSvc = issueSvc

	isGuest, err := ca.CheckUserPermission(ctx)

	if err != nil {
		return err
	}
	ca.isGuest = isGuest

	fixedIssueType := sdk.InParams.String("fixedIssueType")
	if _, ok := issueFilter.CpIssueTypes[fixedIssueType]; !ok {
		return fmt.Errorf("invalid paging request type %v", fixedIssueType)
	}

	projectid, err := strconv.ParseUint(sdk.InParams["projectId"].(string), 10, 64)
	orgid, err := strconv.ParseUint(sdk.Identity.OrgID, 10, 64)

	if err := eventHandler(ctx, event); err != nil {
		return err
	}
	userids := []string{}
	cond := pb.PagingIssueRequest{}
	gh := gshelper.NewGSHelper(gs)
	filterCond, ok := gh.GetIssuePagingRequest()
	if ok {
		cond = *filterCond
		resetPageInfo(&cond, c.State)
	} else {
		issueType := sdk.InParams["fixedIssueType"].(string)
		if _, ok := issueFilter.CpIssueTypes[issueType]; ok {
			if issueType == "ALL" {
				cond.Type = []string{pb.IssueTypeEnum_BUG.String(), pb.IssueTypeEnum_REQUIREMENT.String(), pb.IssueTypeEnum_TASK.String()}
			} else {
				cond.Type = []string{issueType}
			}
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
		cond.IdentityInfo.UserID = sdk.Identity.UserID
	}
	if event.Operation.String() == "changePageNo" {
		cond.PageNo = 1
		if c.State != nil {
			if _, ok := c.State["pageNo"]; ok {
				cond.PageNo = uint64(c.State["pageNo"].(float64))
			}
		}
		cond.PageSize = uint64(c.State["pageSize"].(float64))
		if _, ok := (*gs)["OperationKey"]; ok {
			(*gs)["OperationKey"] = ""
		}
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
		c.State["selectedRowKeys"] = []string{}
	} else if event.Operation.String() == "delete" {
		cputil.MustObjJSONTransfer(&c.State, &ca.state)
		(*gs)["OperationKey"] = "delete"
		(*gs)["SelectedRowKeys"] = ca.state.SelectedRowKeys
	}

	// check reset pageNo
	if c.State != nil && resetPageNoByFilterCondition(event.Operation.String(), TableItem{}, c.State) {
		if _, ok := c.State["pageNo"]; ok {
			cond.PageNo = uint64(c.State["pageNo"].(float64))
		}
	}

	// handle custom properties
	if fixedIssueType != "ALL" {
		cond.WithCustomProperties = true
	}

	issues, total, err := issueQuery.Paging(cond)
	if err != nil {
		return err
	}
	// if pageTotal < cond.PageNo, to reset the cond.PageNo = 1,
	// and return the first page of data
	pageTotal := getTotalPage(total, cond.PageSize)
	if pageTotal < cond.PageNo {
		cond.PageNo = 1
		issues, total, err = issueQuery.Paging(cond)
		if err != nil {
			return err
		}
	}

	for _, p := range issues {
		userids = append(userids, p.Creator, p.Assignee, p.Owner)
		// get from custom properties
		for _, property := range p.PropertyInstances {
			tryUserID := (*issuemodel.PropertyInstanceForShow)(property).TryGetUserID()
			if tryUserID != "" {
				userids = append(userids, tryUserID)
			}
		}
	}
	// 获取全部用户
	userids = strutil.DedupSlice(userids, true)
	if len(issues) > 0 {
		resp, err := identity.FindUsers(ctx, &userpb.FindUsersRequest{IDs: userids})
		if err != nil {
			return err
		}
		ca.userMap = make(map[string]string, len(resp.Data))
		for _, i := range resp.Data {
			ca.userMap[i.ID] = i.Nick
		}
	}

	labels, err := bdl.Bdl.Labels("issue", projectid, sdk.Identity.UserID)
	if err != nil {
		return err
	}
	ca.labels = labels.List

	iterations, _ := gh.GetIterationOptions()
	iterationTitleMap := make(map[int64]string)
	for _, i := range iterations {
		key := int64(i.Value.(float64))
		iterationTitleMap[key] = i.Label
	}

	var l []TableItemForShow
	for _, data := range issues {
		l = append(l, (*ca.buildTableItem(ctx, data, iterationTitleMap)).toColumnDataRow())
	}
	c.Data = map[string]interface{}{}
	c.Data["list"] = l
	var customProperties []*pb.IssuePropertyExtraProperty
	for _, v := range issues {
		if v.PropertyInstances != nil {
			customProperties = v.PropertyInstances
			break
		}
	}
	c.Props = buildTableColumnProps(ctx, fixedIssueType, customProperties)
	c.Operations = map[string]interface{}{
		"changePageNo": map[string]interface{}{
			"key":    "changePageNo",
			"reload": true,
		},
		"changePageSize": map[string]interface{}{
			"key":    "changePageSize",
			"reload": true,
		},
		"delete": map[string]interface{}{
			"confirm": "",
			"key":     "delete",
			"reload":  true,
			"text":    sdk.I18n("Delete"),
		},
	}
	(*gs)[protocol.GlobalInnerKeyUserIDs.String()] = userids
	if c.State == nil {
		c.State = map[string]interface{}{}
	}
	c.State["total"] = total
	c.State["pageNo"] = cond.PageNo
	c.State["pageSize"] = cond.PageSize
	urlquery := fmt.Sprintf(`{"pageNo":%d, "pageSize":%d}`, cond.PageNo, cond.PageSize)
	c.State["issueTable__urlQuery"] = base64.StdEncoding.EncodeToString([]byte(urlquery))
	return nil
}

func (ca *ComponentAction) buildTableItem(ctx context.Context, data *pb.Issue, iterations map[int64]string) *TableItem {
	var issuestate *pb.IssueStateButton
	for _, s := range data.IssueButton {
		if s.StateID == data.State {
			issuestate = s
			break
		}
	}
	nameColumn := ca.getNameColumn(data)
	progress := Progress{
		RenderType: "multiple",
		Direction:  "row",
	}
	if data.Type == pb.IssueTypeEnum_REQUIREMENT {
		if data.IssueSummary == nil {
			data.IssueSummary = &pb.IssueSummary{}
		}
		s := data.IssueSummary.DoneCount + data.IssueSummary.ProcessingCount
		progressPercentage := ProgressBlock{
			RenderType: "progress",
			Value:      "0",
			HiddenText: true,
		}
		if s != 0 {
			progressPercentage.Value = fmt.Sprintf("%d", int(100*(float64(data.IssueSummary.DoneCount)/float64(s))))
		}
		progress.Renders = []interface{}{
			[]interface{}{
				progressPercentage,
			},
			[]interface{}{
				ProgressBlock{
					RenderType: "text",
					Value:      fmt.Sprintf("%d/%d", data.IssueSummary.DoneCount, s),
				},
			},
		}
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
			"text":       cputil.I18n(ctx, s.GetI18nKeyAlias()),
			"prefixIcon": "ISSUE_ICON.severity." + string(s),
			"meta": map[string]string{
				"id":       strconv.FormatInt(data.Id, 10),
				"severity": string(s),
			},
		}
	}
	severity := Severity{
		RenderType: "operationsDropdownMenu",
		Value:      cputil.I18n(ctx, GetI18nKeyAlias(data.Severity)),
		PrefixIcon: "ISSUE_ICON.severity." + data.Severity.String(),
		Operations: severityOps,
		Disabled:   ca.isGuest,
		DisabledTip: map[bool]string{
			true: "无权限",
		}[ca.isGuest],
	}
	stateMenus := make([]map[string]interface{}, 0, len(data.IssueButton))
	stateAllDisable := true
	for i, s := range data.IssueButton {
		if s.Permission {
			stateAllDisable = false
		}
		if ca.isGuest {
			stateAllDisable = true
		}
		menu := map[string]interface{}{
			"meta": map[string]string{
				"state": strconv.FormatInt(s.StateID, 10),
				"id":    strconv.FormatInt(data.Id, 10),
			},
			"id": s.StateName,
			// "prefixIcon": stateIcon[string(s.StateBelong)],
			"status":   common.GetUIIssueState(apistructs.IssueStateBelong(s.StateBelong.String())),
			"text":     s.StateName,
			"reload":   true,
			"key":      "changeStateTo" + strconv.Itoa(i) + s.StateName,
			"disabled": !s.Permission,
			"disabledTip": map[bool]string{
				false: "无法转移",
			}[s.Permission],
		}
		if !s.Permission {
			menu["hidden"] = true
		}
		stateMenus = append(stateMenus, menu)
	}
	AssigneeMapOperations := map[string]interface{}{}
	AssigneeMapOperations["onChange"] = map[string]interface{}{
		"meta": map[string]string{
			"assignee": "",
			"id":       strconv.FormatInt(data.Id, 10),
		},
		"text":     ca.userMap[data.Assignee],
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
					"id":            strconv.FormatInt(data.Id, 10),
					"deadlineValue": "",
				},
			},
		},
	}
	if data.PlanFinishedAt != nil {
		deadline.Value = data.PlanFinishedAt.AsTime().Format(time.RFC3339)
	}
	if data.PlanStartedAt != nil {
		deadline.DisabledBefore = data.PlanStartedAt.AsTime().Format(time.RFC3339)
	}
	closedAt := buildTime(data.FinishTime)
	createdAt := buildTime(data.CreatedAt)
	planStartedAt := buildTime(data.PlanStartedAt)
	state := State{
		// Operations: stateOperations,
		RenderType: "dropdownMenu",
		Menus:      stateMenus,
		Disabled:   stateAllDisable,
		DisabledTip: map[bool]string{
			true: "无权限",
		}[stateAllDisable],
	}
	if issuestate != nil {
		// state.PrefixIcon = stateIcon[string(issuestate.StateBelong)]
		state.Value = issuestate.StateName
	}
	iteration := TextBlock{
		RenderType: "text",
		Value:      "",
	}

	if iterations != nil {
		if v, ok := iterations[data.IterationID]; ok {
			iteration.Value = v
		}
	}
	return &TableItem{
		//Assignee:    map[string]string{"value": data.Assignee, "renderType": "userAvatar"},
		Id:          strconv.FormatInt(data.Id, 10),
		IterationID: data.IterationID,
		Type:        data.Type.String(),
		Progress:    progress,
		Severity:    severity,
		Complexity: Complexity{
			RenderType: "textWithIcon",
			PrefixIcon: "ISSUE_ICON.complexity." + data.Complexity.String(),
			Value:      cputil.I18n(ctx, data.Complexity.String()),
		},
		Priority: Priority{
			Value:      cputil.I18n(ctx, strings.ToLower(data.Priority.String())),
			RenderType: "operationsDropdownMenu",
			PrefixIcon: priorityIcon[apistructs.IssuePriority(data.Priority.String())],
			Operations: map[string]interface{}{
				"changePriorityTodLOW": map[string]interface{}{
					"meta": map[string]string{
						"id":       strconv.FormatInt(data.Id, 10),
						"priority": "LOW",
					},
					"prefixIcon": priorityIcon[apistructs.IssuePriorityLow],
					"text":       cputil.I18n(ctx, "low"),
					"reload":     true,
					"key":        "changePriorityTodLOW",
				}, "changePriorityTocNORMAL": map[string]interface{}{
					"meta": map[string]string{
						"id":       strconv.FormatInt(data.Id, 10),
						"priority": "NORMAL",
					},
					"prefixIcon": priorityIcon[apistructs.IssuePriorityNormal],
					"text":       cputil.I18n(ctx, "normal"),
					"reload":     true,
					"key":        "changePriorityTocNORMAL",
				}, "changePriorityTobHIGH": map[string]interface{}{
					"meta": map[string]string{
						"id":       strconv.FormatInt(data.Id, 10),
						"priority": "HIGH",
					},
					"prefixIcon": priorityIcon[apistructs.IssuePriorityHigh],
					"text":       cputil.I18n(ctx, "high"),
					"reload":     true,
					"key":        "changePriorityTobHIGH",
				},
				"changePriorityToaURGENT": map[string]interface{}{
					"meta": map[string]string{
						"id":       strconv.FormatInt(data.Id, 10),
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
		Name:      nameColumn,
		Deadline:  deadline,
		ClosedAt:  closedAt,
		CreatedAt: createdAt,
		ReopenCount: TextBlock{
			RenderType: "text",
			Value:      fmt.Sprintf("%d", data.ReopenCount),
		},
		Creator: Assignee{
			Value:      data.Creator,
			RenderType: "userAvatar",
		},
		Owner: Assignee{
			Value:      data.Owner,
			RenderType: "userAvatar",
		},
		PlanStartedAt: planStartedAt,
		Iteration:     iteration,
		Properties:    data.PropertyInstances,
	}
}

func (ca *ComponentAction) getNameColumn(issue *pb.Issue) Name {
	var tags []Label
	for _, label := range issue.Labels {
		for _, labelDef := range ca.labels {
			if label == labelDef.Name {
				tags = append(tags, Label{Color: labelDef.Color, Label: labelDef.Name})
			}
		}
	}
	return Name{
		RenderType: "doubleRowWithIcon",
		PrefixIcon: getPrefixIcon(issue.Type.String()),
		Value:      issue.Title,
		ExtraContent: ExtraContent{
			RenderType: "tags",
			Value:      tags,
			ShowCount:  4,
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

func resetPageInfo(req *pb.PagingIssueRequest, state map[string]interface{}) {
	req.PageSize = 10
	if _, ok := state["pageNo"]; ok {
		req.PageNo = uint64(state["pageNo"].(float64))
	}
	if _, ok := state["pageSize"]; ok {
		req.PageSize = uint64(state["pageSize"].(float64))
	}
}

func GetI18nKeyAlias(is pb.IssueSeverityEnum_Severity) string {
	if is == pb.IssueSeverityEnum_NORMAL {
		return "ordinary"
	}
	return strings.ToLower(is.String())
}

func buildTime(t *timestamppb.Timestamp) Time {
	res := Time{
		RenderType: "datePicker",
		Value:      "",
		NoBorder:   true,
	}
	if t != nil {
		res.Value = t.AsTime().Format(time.RFC3339)
	}
	return res
}

func eventHandler(ctx context.Context, event cptype.ComponentEvent) error {
	sdk := cputil.SDK(ctx)
	issueSvc := ctx.Value(types.IssueQuery).(query.Interface)
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
		p := string(priority)
		err = issueSvc.UpdateIssue(&pb.UpdateIssueRequest{
			Id:       id,
			Priority: &p,
			IdentityInfo: &commonpb.IdentityInfo{
				UserID: sdk.Identity.UserID,
			},
		})
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
		err = issueSvc.UpdateIssue(&pb.UpdateIssueRequest{
			Id:             id,
			PlanFinishedAt: &operationData.Meta.DeadlineValue,
			IdentityInfo: &commonpb.IdentityInfo{
				UserID: sdk.Identity.UserID,
			},
		})
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
		err = issueSvc.UpdateIssue(&pb.UpdateIssueRequest{
			Id:    id,
			State: &stateid,
			IdentityInfo: &commonpb.IdentityInfo{
				UserID: sdk.Identity.UserID,
			},
		})
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
		severity := operationData.Meta.Severity
		err = issueSvc.UpdateIssue(&pb.UpdateIssueRequest{
			Id:       id,
			Severity: &severity,
			IdentityInfo: &commonpb.IdentityInfo{
				UserID: sdk.Identity.UserID,
			},
		})
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
		assignee := operationData.Meta.Assignee
		err = issueSvc.UpdateIssue(&pb.UpdateIssueRequest{
			Id:       id,
			Assignee: &assignee,
			IdentityInfo: &commonpb.IdentityInfo{
				UserID: sdk.Identity.UserID,
			},
		})
		if err != nil {
			logrus.Errorf("update issue failed, id:%v, err:%v", id, err)
			return err
		}
	}
	return nil
}
