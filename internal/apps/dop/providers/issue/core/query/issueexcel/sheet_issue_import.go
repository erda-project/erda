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

package issueexcel

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	issuedao "github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
)

func (data DataForFulfill) DecodeIssueSheet(excelSheets [][][]string) ([]IssueSheetModel, error) {
	sheet := excelSheets[indexOfSheetIssue]
	if data.IsOldExcelFormat() {
		return data.convertOldIssueSheet(sheet)
	}
	// convert [][][]string to map[uuid]excel.Column
	issueSheetRows := sheet
	var columnIndex int
	for _, row := range issueSheetRows {
		columnIndex = len(row)
		break
	}
	if columnIndex == 0 {
		return nil, fmt.Errorf("invalid issue sheet, no column")
	}
	m := make(map[IssueSheetColumnUUID]excel.Column)
	for i := 0; i < columnIndex; i++ {
		// format like "Common---ID---ID"
		var uuid IssueSheetColumnUUID
		for j := 0; j < uuidPartsMustLength; j++ {
			uuid.AddPart(issueSheetRows[j][i])
		}
		// data rows start from uuidPartsMustLength
		for k := range issueSheetRows[uuidPartsMustLength:] {
			row := issueSheetRows[uuidPartsMustLength+k]
			m[uuid] = append(m[uuid], excel.Cell{
				Value:              row[i],
				VerticalMergeNum:   0,
				HorizontalMergeNum: 0,
			})
		}
	}
	// decode map to models
	models, err := data.decodeMapToIssueSheetModel(m)
	if err != nil {
		return nil, fmt.Errorf("failed to decode map to models, err: %v", err)
	}
	return models, nil
}

func (data DataForFulfill) decodeMapToIssueSheetModel(m map[IssueSheetColumnUUID]excel.Column) (_ []IssueSheetModel, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			// print runtime stacktrace
			fmt.Println(string(debug.Stack()))
		}
	}()
	// prepare models
	var dataRowsNum int
	for _, column := range m {
		dataRowsNum = len(column)
		break
	}
	models := make([]IssueSheetModel, dataRowsNum, dataRowsNum)
	for uuid, column := range m {
		parts := uuid.Decode()
		if len(parts) != uuidPartsMustLength {
			return nil, fmt.Errorf("invalid uuid: %s", uuid)
		}
		for i, cell := range column {
			model := &models[i]
			groupType := parts[0]
			groupField := parts[1]
			switch groupType {
			case "Common":
				switch groupField {
				case "ID":
					if cell.Value == "" { // update
						model.Common.ID = 0
						continue
					}
					id, err := strconv.ParseUint(cell.Value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("invalid id: %s", cell.Value)
					}
					model.Common.ID = id
				case "IterationName":
					model.Common.IterationName = cell.Value
				case "IssueType":
					issueType, err := parseStringIssueType(cell.Value)
					if err != nil {
						return nil, err
					}
					model.Common.IssueType = issueType
				case "IssueTitle":
					model.Common.IssueTitle = cell.Value
				case "Content":
					model.Common.Content = cell.Value
				case "State":
					model.Common.State = cell.Value
				case "Priority":
					priority, err := parseStringPriority(cell.Value)
					if err != nil {
						return nil, err
					}
					model.Common.Priority = priority
				case "Complexity":
					complexity, err := parseStringComplexity(cell.Value)
					if err != nil {
						return nil, err
					}
					model.Common.Complexity = complexity
				case "Severity":
					severity, err := parseStringSeverity(cell.Value)
					if err != nil {
						return nil, err
					}
					model.Common.Severity = severity
				case "CreatorName":
					model.Common.CreatorName = cell.Value
				case "AssigneeName":
					model.Common.AssigneeName = cell.Value
				case "CreatedAt":
					model.Common.CreatedAt = mustParseStringTime(cell.Value, groupField)
				case "UpdatedAt":
					model.Common.UpdatedAt = mustParseStringTime(cell.Value, groupField)
				case "PlanStartedAt":
					model.Common.PlanStartedAt = mustParseStringTime(cell.Value, groupField)
				case "PlanFinishedAt":
					model.Common.PlanFinishedAt = mustParseStringTime(cell.Value, groupField)
				case "StartAt":
					model.Common.StartAt = mustParseStringTime(cell.Value, groupField)
				case "FinishAt":
					model.Common.FinishAt = mustParseStringTime(cell.Value, groupField)
				case "EstimateTime":
					model.Common.EstimateTime = cell.Value
				case "Labels":
					model.Common.Labels = parseStringSliceByComma(cell.Value)
				case "ConnectionIssueIDs":
					var ids []int64
					for _, idStr := range parseStringSliceByComma(cell.Value) {
						id, err := parseStringIssueID(idStr)
						if err != nil {
							return nil, fmt.Errorf("invalid connection issue id: %s", idStr)
						}
						if id != nil {
							ids = append(ids, *id)
						}
					}
					model.Common.ConnectionIssueIDs = ids
				default:
					return nil, fmt.Errorf("unknown common field: %s", groupField)
				}
			case "RequirementOnly":
				switch groupField {
				case "InclusionIssueIDs":
					var ids []int64
					for _, idStr := range parseStringSliceByComma(cell.Value) {
						id, err := parseStringIssueID(idStr)
						if err != nil {
							return nil, fmt.Errorf("invalid inclusion issue id: %s", idStr)
						}
						if id != nil {
							ids = append(ids, *id)
						}
					}
					model.RequirementOnly.InclusionIssueIDs = ids
				case "CustomFields":
					model.RequirementOnly.CustomFields = append(model.RequirementOnly.CustomFields, ExcelCustomField{
						Title: parts[2],
						Value: cell.Value,
					})
				}
			case "TaskOnly":
				switch groupField {
				case "TaskType":
					model.TaskOnly.TaskType = cell.Value
				case "CustomFields":
					model.TaskOnly.CustomFields = append(model.TaskOnly.CustomFields, ExcelCustomField{
						Title: parts[2],
						Value: cell.Value,
					})
				}
			case "BugOnly":
				switch groupField {
				case "OwnerName":
					model.BugOnly.OwnerName = cell.Value
				case "Source":
					model.BugOnly.Source = cell.Value
				case "ReopenCount":
					if cell.Value == "" {
						model.BugOnly.ReopenCount = 0
						continue
					}
					reopenCount, err := strconv.ParseInt(cell.Value, 10, 32)
					if err != nil {
						return nil, fmt.Errorf("invalid reopen count: %s", cell.Value)
					}
					model.BugOnly.ReopenCount = int32(reopenCount)
				case "CustomFields":
					model.BugOnly.CustomFields = append(model.BugOnly.CustomFields, ExcelCustomField{
						Title: parts[2],
						Value: cell.Value,
					})
				}
			}
		}

	}
	return models, nil
}

func parseStringTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func mustParseStringTime(s string, typ string) *time.Time {
	t, err := parseStringTime(s)
	if err != nil {
		panic(fmt.Sprintf("invalid %s time: %s, err: %v", typ, s, err))
	}
	return t
}

func parseStringSliceByComma(s string) []string {
	results := strutil.Splits(s, []string{",", "，"}, true)
	// trim space
	for i, v := range results {
		results[i] = strings.TrimSpace(v)
	}
	return results
}

// createOrUpdateIssues 创建或更新 issues
// 根据 project id 进行判断
func (data DataForFulfill) createOrUpdateIssues(issueSheetModels []IssueSheetModel) (_ []*issuedao.Issue, issueModelMapByIssueID map[uint64]*IssueSheetModel, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			fmt.Println(string(debug.Stack()))
		}
	}()
	idMapping := make(map[int64]uint64) // key: old id(can be negative for Excel Line Num, like L5), value: new id
	issueModelMapByIssueID = make(map[uint64]*IssueSheetModel)
	var issues []*issuedao.Issue
	for i, model := range issueSheetModels {
		model := model
		// check state
		stateID, ok := data.StateMapByTypeAndName[model.Common.IssueType.String()][model.Common.State]
		if !ok {
			return nil, nil, fmt.Errorf("unknown state: %s, please contact project manager to add the corresponding status first", model.Common.State)
		}
		issue := issuedao.Issue{
			BaseModel: dbengine.BaseModel{
				ID:        uint64(model.Common.ID),
				CreatedAt: changePointerTimeToTime(model.Common.CreatedAt),
				UpdatedAt: changePointerTimeToTime(model.Common.UpdatedAt),
			},
			PlanStartedAt:  model.Common.PlanStartedAt,
			PlanFinishedAt: model.Common.PlanFinishedAt,
			ProjectID:      data.ProjectID,
			IterationID:    int64(data.IterationMapByName[model.Common.IterationName].ID),
			AppID:          nil,
			RequirementID:  nil,
			Type:           model.Common.IssueType.String(),
			Title:          model.Common.IssueTitle,
			Content:        model.Common.Content,
			State:          stateID,
			Priority:       model.Common.Priority.String(),
			Complexity:     model.Common.Complexity.String(),
			Severity:       model.Common.Severity.String(),
			Creator:        data.ImportOnly.UserIDByNick[model.Common.CreatorName],
			Assignee:       data.ImportOnly.UserIDByNick[model.Common.AssigneeName],
			Source:         "",
			ManHour:        mustGetJsonManHour(model.Common.EstimateTime),
			External:       true,
			Deleted:        false,
			Stage:          getIssueStage(model),
			Owner:          data.ImportOnly.UserIDByNick[model.BugOnly.OwnerName],
			FinishTime:     model.Common.FinishAt,
			ExpiryStatus:   "",
			ReopenCount:    int(model.BugOnly.ReopenCount),
			StartTime:      model.Common.StartAt,
		}
		if issue.ID > 0 && data.ShouldUpdateWhenIDSame() && data.ImportOnly.CurrentProjectIssueMap[issue.ID] {
			// update
			if err := data.ImportOnly.DB.UpdateIssueType(&issue); err != nil {
				return nil, nil, fmt.Errorf("failed to update issue, id: %d, err: %v", issue.ID, err)
			}
		} else {
			// create
			issue.ID = 0
			if err := data.ImportOnly.DB.CreateIssue(&issue); err != nil {
				return nil, nil, fmt.Errorf("failed to create issue, err: %v", err)
			}
		}
		issues = append(issues, &issue)
		issueModelMapByIssueID[issue.ID] = &model
		// got new issue id here, set id mapping
		idMapping[int64(model.Common.ID)] = issue.ID
		idMapping[-int64(i+(uuidPartsMustLength+1))] = issue.ID // excel line num, star from 1
	}
	// all id mapping done here, update old ids
	for i := range issueSheetModels {
		model := issueSheetModels[i]
		// id
		model.Common.ID = idMapping[int64(model.Common.ID)]
		// connection issue id
		for j := range model.Common.ConnectionIssueIDs {
			newID, ok := idMapping[model.Common.ConnectionIssueIDs[j]]
			if !ok {
				return nil, nil, fmt.Errorf("invalid connection issue id: %d", model.Common.ConnectionIssueIDs[j])
			}
			model.Common.ConnectionIssueIDs[j] = int64(newID)
		}
		// inclusion issue id
		for j := range model.RequirementOnly.InclusionIssueIDs {
			newID, ok := idMapping[model.RequirementOnly.InclusionIssueIDs[j]]
			if !ok {
				return nil, nil, fmt.Errorf("invalid inclusion issue id: %d", model.RequirementOnly.InclusionIssueIDs[j])
			}
			model.RequirementOnly.InclusionIssueIDs[j] = int64(newID)
		}
		issueSheetModels[i] = model
		issueModelMapByIssueID[model.Common.ID] = &model
	}
	return issues, issueModelMapByIssueID, nil
}

func (data DataForFulfill) createIssueRelations(issues []*issuedao.Issue, issueModelMapByIssueID map[uint64]*IssueSheetModel) error {
	for _, issue := range issues {
		model, ok := issueModelMapByIssueID[issue.ID]
		if !ok {
			return fmt.Errorf("cannot find issue model by issue id: %d", issue.ID)
		}
		// connections
		var issueRelations []issuedao.IssueRelation
		for _, connectionIssueID := range model.Common.ConnectionIssueIDs {
			issueRelations = append(issueRelations, issuedao.IssueRelation{
				IssueID:      issue.ID,
				RelatedIssue: uint64(connectionIssueID),
				Comment:      "",
				Type:         apistructs.IssueRelationConnection,
			})
		}
		// inclusions
		for _, inclusionIssueID := range model.RequirementOnly.InclusionIssueIDs {
			issueRelations = append(issueRelations, issuedao.IssueRelation{
				IssueID:      issue.ID,
				RelatedIssue: uint64(inclusionIssueID),
				Comment:      "",
				Type:         apistructs.IssueRelationInclusion,
			})
		}
		// insert into db
		if err := data.ImportOnly.DB.BatchCreateIssueRelations(issueRelations); err != nil {
			return fmt.Errorf("failed to batch create issue relations, err: %v", err)
		}
	}
	return nil
}

// parseStringIssueID
// format:
// - normal: 123
// - line num: L5
func parseStringIssueID(s string) (*int64, error) {
	if s == "" {
		return nil, nil
	}
	// line num, like: L5
	if strings.HasPrefix(s, "L") {
		s = s[1:]
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid string line id: %s", s)
		}
		minLen := uuidPartsMustLength + 1
		if i < int64(minLen) {
			return nil, fmt.Errorf("invalid string line id: %s, line num cannot lower than %d", s, minLen)
		}
		i = -i
		return &i, nil
	}
	// normal
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid string line id: %s", s)
	}
	if i < 0 {
		return nil, fmt.Errorf("invalid string line id: %s, cannot lower than 1", s)
	}
	return &i, nil
}

func parseStringIssueType(s string) (pb.IssueTypeEnum_Type, error) {
	var t pb.IssueTypeEnum_Type
	switch strings.ToLower(s) {
	case strings.ToLower(pb.IssueTypeEnum_REQUIREMENT.String()), "需求":
		t = pb.IssueTypeEnum_REQUIREMENT
	case strings.ToLower(pb.IssueTypeEnum_TASK.String()), "任务":
		t = pb.IssueTypeEnum_TASK
	case strings.ToLower(pb.IssueTypeEnum_BUG.String()), "缺陷":
		t = pb.IssueTypeEnum_BUG
	case strings.ToLower(pb.IssueTypeEnum_EPIC.String()), "史诗":
		t = pb.IssueTypeEnum_EPIC
	case strings.ToLower(pb.IssueTypeEnum_TICKET.String()), "工单":
		t = pb.IssueTypeEnum_TICKET
	default:
		return t, fmt.Errorf("invalid issue type: %s", s)
	}
	return t, nil
}

func parseStringPriority(s string) (pb.IssuePriorityEnum_Priority, error) {
	var p pb.IssuePriorityEnum_Priority
	switch strings.ToLower(s) {
	case strings.ToLower(pb.IssuePriorityEnum_LOW.String()), "低":
		p = pb.IssuePriorityEnum_LOW
	case strings.ToLower(pb.IssuePriorityEnum_NORMAL.String()), "中":
		p = pb.IssuePriorityEnum_NORMAL
	case strings.ToLower(pb.IssuePriorityEnum_HIGH.String()), "高":
		p = pb.IssuePriorityEnum_HIGH
	case strings.ToLower(pb.IssuePriorityEnum_URGENT.String()), "紧急":
		p = pb.IssuePriorityEnum_URGENT
	default:
		return p, fmt.Errorf("invalid issue priority: %s", s)
	}
	return p, nil
}

func parseStringComplexity(s string) (pb.IssueComplexityEnum_Complextity, error) {
	var c pb.IssueComplexityEnum_Complextity
	switch strings.ToLower(s) {
	case strings.ToLower(pb.IssueComplexityEnum_EASY.String()), "容易":
		c = pb.IssueComplexityEnum_EASY
	case strings.ToLower(pb.IssueComplexityEnum_NORMAL.String()), "中":
		c = pb.IssueComplexityEnum_NORMAL
	case strings.ToLower(pb.IssueComplexityEnum_HARD.String()), "复杂":
		c = pb.IssueComplexityEnum_HARD
	default:
		return c, fmt.Errorf("invalid issue complexity: %s", s)
	}
	return c, nil
}

func parseStringSeverity(s string) (pb.IssueSeverityEnum_Severity, error) {
	var c pb.IssueSeverityEnum_Severity
	switch strings.ToLower(s) {
	case strings.ToLower(pb.IssueSeverityEnum_FATAL.String()), "致命":
		c = pb.IssueSeverityEnum_FATAL
	case strings.ToLower(pb.IssueSeverityEnum_SERIOUS.String()), "严重":
		c = pb.IssueSeverityEnum_SERIOUS
	case strings.ToLower(pb.IssueSeverityEnum_NORMAL.String()), "一般":
		c = pb.IssueSeverityEnum_NORMAL
	case strings.ToLower(pb.IssueSeverityEnum_SLIGHT.String()), "轻微":
		c = pb.IssueSeverityEnum_SLIGHT
	case strings.ToLower(pb.IssueSeverityEnum_SUGGEST.String()), "建议":
		c = pb.IssueSeverityEnum_SUGGEST
	default:
		return c, fmt.Errorf("invalid issue severity: %s", s)
	}
	return c, nil
}
