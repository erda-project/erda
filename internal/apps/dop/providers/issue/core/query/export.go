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

package query

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/common"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	streamcommon "github.com/erda-project/erda/internal/apps/dop/providers/issue/stream/common"
	"github.com/erda-project/erda/internal/apps/dop/services/i18n"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
)

func (p *provider) ExportExcel(issues []*pb.Issue, properties []*pb.IssuePropertyIndex, projectID uint64, isDownloadTemplate bool, orgID int64, locale string) (io.Reader, string, error) {
	// list of issue stage
	stages, err := p.db.GetIssuesStageByOrgID(orgID)
	if err != nil {
		return nil, "", err
	}
	// get the stageMap
	stageMap := getStageMap(stages)

	table, err := p.convertIssueToExcelList(issues, properties, projectID, isDownloadTemplate, stageMap, locale)
	if err != nil {
		return nil, "", err
	}
	// replace userids by usernames
	userids := []string{}
	for _, t := range table[1:] {
		if t[4] != "" {
			userids = append(userids, t[4])
		}
		if t[5] != "" {
			userids = append(userids, t[5])
		}
		if t[6] != "" {
			userids = append(userids, t[6])
		}
	}
	userids = strutil.DedupSlice(userids, true)
	resp, err := p.Identity.FindUsers(context.Background(), &userpb.FindUsersRequest{
		IDs: userids,
	})
	users := resp.Data
	if err != nil {
		return nil, "", err
	}
	usernames := map[string]string{}
	for _, u := range users {
		usernames[u.ID] = u.Nick
	}
	for i := 1; i < len(table); i++ {
		if table[i][4] != "" {
			if name, ok := usernames[table[i][4]]; ok {
				table[i][4] = name
			}
		}
		if table[i][5] != "" {
			if name, ok := usernames[table[i][5]]; ok {
				table[i][5] = name
			}
		}
		if table[i][6] != "" {
			if name, ok := usernames[table[i][6]]; ok {
				table[i][6] = name
			}
		}
	}
	tablename := "issuetable"
	if len(issues) > 0 {
		if issues[0].IterationID == -1 {
			tablename = "待办事项"
		} else {
			tablename = issues[0].Type.String()
		}
	}

	// insert sample issue
	if isDownloadTemplate {
		table = append(table, p.getIssueExportDataI18n(locale, i18n.I18nKeyIssueExportSample))
	}
	buf := bytes.NewBuffer([]byte{})
	if err := excel.ExportExcel(buf, table, tablename); err != nil {
		return nil, "", err
	}
	return buf, tablename, nil
}

// getStageMap return a map,the key is the struct of dice_issue_stage.Value and dice_issue_stage.IssueType,
// the value is dice_issue_stage.Name
func getStageMap(stages []dao.IssueStage) map[IssueStage]string {
	stageMap := make(map[IssueStage]string, len(stages))
	for _, v := range stages {
		if v.Value != "" && v.IssueType != "" {
			stage := IssueStage{
				Type:  v.IssueType,
				Value: v.Value,
			}
			stageMap[stage] = v.Name
		}
	}
	return stageMap
}

// convertIssueToExcelList convert issue to excel list
func (p *provider) convertIssueToExcelList(issues []*pb.Issue, property []*pb.IssuePropertyIndex, projectID uint64, isDownloadTemplate bool, stageMap map[IssueStage]string, locale string) ([][]string, error) {
	// 默认字段列名
	r := [][]string{p.getIssueExportDataI18n(locale, i18n.I18nKeyIssueExportTitles)}
	//var excelRows [][]excel.Cell
	//// common title
	//titleCommonStrs := p.getIssueExportDataI18n(locale, i18n.I18nKeyIssueExportTitleCommon)
	//titleCommonCells := excel.ConvertStringSliceToCellSlice(titleCommonStrs)
	//// requirement-only title
	//titleRequirementOnlyStrs := p.getIssueExportDataI18n(locale, i18n.I18nKeyIssueExportTitleRequirementOnly)

	// 自定义字段列名
	for _, pro := range property {
		r[0] = append(r[0], pro.DisplayName)
	}
	// 下载模版
	if isDownloadTemplate {
		return r, nil
	}
	// 构建自定义字段枚举值map
	mp := make(map[pair]string)
	for _, v := range property {
		if common.IsOptions(v.PropertyType.String()) == true {
			for _, val := range v.EnumeratedValues {
				mp[pair{PropertyID: v.PropertyID, valueID: val.Id}] = val.Name
			}
		}
	}
	// 状态名map
	stateMap := make(map[int64]string)
	states, err := p.db.GetIssuesStatesByProjectID(projectID, "")
	if err != nil {
		return nil, err
	}
	for _, v := range states {
		stateMap[int64(v.ID)] = v.Name
	}
	// 迭代map
	iterationMap := make(map[int64]string)
	iterations, _, err := p.db.PagingIterations(apistructs.IterationPagingRequest{
		PageNo:    1,
		PageSize:  10000,
		ProjectID: projectID,
	})
	if err != nil {
		return nil, err
	}
	for _, v := range iterations {
		iterationMap[int64(v.ID)] = v.Title
	}
	iterationMap[-1] = "待办事项"
	// 自定义字段map
	var issueIDs []int64
	for _, v := range issues {
		issueIDs = append(issueIDs, v.Id)
	}
	propertyMap := make(map[int64][]dao.IssuePropertyRelation)
	properties, err := p.db.PagingPropertyRelationByIDs(issueIDs)
	if err != nil {
		return nil, err
	}
	userIDs := make([]string, 0)
	for _, v := range properties {
		for _, pro := range property {
			if pro.PropertyID == v.PropertyID && pro.PropertyType == pb.PropertyTypeEnum_Person {
				userIDs = append(userIDs, v.ArbitraryValue)
			}
		}
		propertyMap[v.IssueID] = append(propertyMap[v.IssueID], v)
	}
	userIDs = strutil.DedupSlice(userIDs, true)
	usernames := map[string]string{}
	if len(userIDs) > 0 {
		resp, err := p.Identity.FindUsers(context.Background(), &userpb.FindUsersRequest{
			IDs: userIDs,
		})
		users := resp.Data
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			usernames[u.ID] = u.Nick
		}
	}
	for index, i := range issues {
		planFinishedAt := ""
		if i.PlanFinishedAt != nil {
			planFinishedAt = i.PlanFinishedAt.AsTime().In(time.Local).Format("2006-01-02 15:04:05")
		}
		planStartedAt := ""
		if i.PlanStartedAt != nil {
			planStartedAt = i.PlanStartedAt.AsTime().In(time.Local).Format("2006-01-02 15:04:05")
		}
		iterationName := iterationMap[i.IterationID]
		stage := IssueStage{
			Type:  i.Type.String(),
			Value: common.GetStage(i),
		}
		finishTime := ""
		if i.FinishTime != nil {
			finishTime = i.FinishTime.AsTime().In(time.Local).Format("2006-01-02 15:04:05")
		}

		_, relatedIssueIDs, err := p.GetIssueRelationsByIssueIDs(uint64(i.Id), []string{apistructs.IssueRelationConnection})
		if err != nil {
			return nil, err
		}
		relatedIssueIDStrs := make([]string, 0)
		for _, id := range relatedIssueIDs {
			relatedIssueIDStrs = append(relatedIssueIDStrs, strconv.FormatUint(id, 10))
		}

		r = append(r, append([]string{
			strconv.FormatInt(i.Id, 10),
			i.Title,
			i.Content,
			stateMap[i.State],
			i.Creator,
			i.Assignee,
			i.Owner,
			stageMap[stage],
			map[pb.IssuePriorityEnum_Priority]string{
				pb.IssuePriorityEnum_LOW:    "低",
				pb.IssuePriorityEnum_HIGH:   "高",
				pb.IssuePriorityEnum_NORMAL: "中",
				pb.IssuePriorityEnum_URGENT: "紧急",
			}[i.Priority],
			iterationName,
			map[pb.IssueComplexityEnum_Complextity]string{
				pb.IssueComplexityEnum_EASY:   "容易",
				pb.IssueComplexityEnum_NORMAL: "中",
				pb.IssueComplexityEnum_HARD:   "复杂",
			}[i.Complexity],
			map[pb.IssueSeverityEnum_Severity]string{
				pb.IssueSeverityEnum_FATAL:   "致命",
				pb.IssueSeverityEnum_SERIOUS: "严重",
				pb.IssueSeverityEnum_NORMAL:  "一般",
				pb.IssueSeverityEnum_SLIGHT:  "轻微",
				pb.IssueSeverityEnum_SUGGEST: "建议",
			}[i.Severity],
			strutil.Join(i.Labels, ",", true),
			map[pb.IssueTypeEnum_Type]string{
				pb.IssueTypeEnum_REQUIREMENT: "需求",
				pb.IssueTypeEnum_TASK:        "任务",
				pb.IssueTypeEnum_BUG:         "缺陷",
				pb.IssueTypeEnum_EPIC:        "史诗",
				pb.IssueTypeEnum_TICKET:      "工单",
			}[i.Type],
			planFinishedAt,
			i.CreatedAt.AsTime().In(time.Local).Format("2006-01-02 15:04:05"),
			strings.Join(relatedIssueIDStrs, ","),
			streamcommon.GetFormartTime(i.IssueManHour, "EstimateTime"),
			finishTime,
			planStartedAt,
			fmt.Sprintf("%d", i.ReopenCount),
		}))
		relations := propertyMap[i.Id]
		// get value of each custom field
		for _, pro := range property {
			columnValue := getCustomPropertyColumnValue(pro, relations, mp, usernames)
			r[index+1] = append(r[index+1], columnValue)
		}
	}
	return r, nil
}
