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

package sheet_issue

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	issuedao "github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
)

type Handler struct{}

func (h *Handler) SheetName() string { return vars.NameOfSheetIssue }

func (h *Handler) ImportSheet(data *vars.DataForFulfill, s *excel.Sheet) error {
	if data.IsOldExcelFormat() {
		convertedIssueSheetModels, err := convertOldIssueSheet(data, s.UnmergedSlice)
		if err != nil {
			return fmt.Errorf("failed to convert old issue sheet, err: %v", err)
		}
		data.ImportOnly.Sheets.Must.IssueInfo = convertedIssueSheetModels
		return nil
	}
	sheet := s.UnmergedSlice
	// convert [][][]string to map[uuid]excel.Column
	issueSheetRows := sheet
	// polish sheet rows, remove empty rows
	removeEmptySheetRows(&issueSheetRows)
	var columnLen int
	for _, row := range issueSheetRows {
		columnLen = len(row)
		break
	}
	if columnLen == 0 {
		return fmt.Errorf("invalid issue sheet, no column")
	}
	// auto fill empty cells
	autoFillEmptyRowCells(&issueSheetRows, columnLen)

	m := make(map[IssueSheetColumnUUID]excel.Column)
	for i := 0; i < columnLen; i++ {
		// format like "Common---ID---ID"
		var uuid IssueSheetColumnUUID
		for j := 0; j < uuidPartsMustLength; j++ {
			// parse i18n text to excel field key
			rawValue := issueSheetRows[j][i]
			cellValue := parseI18nTextToExcelFieldKey(rawValue)
			// skip concrete custom field
			parts := uuid.Decode()
			if len(parts) >= 2 && parts[1] == fieldCustomFields { // it's the third part and part 2 is `CustomFields`
				cellValue = rawValue
			}
			uuid.AddPart(cellValue)
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
	models, err := decodeMapToIssueSheetModel(data, m)
	if err != nil {
		return fmt.Errorf("failed to decode map to models, err: %v", err)
	}
	data.ImportOnly.Sheets.Must.IssueInfo = models
	return nil
}

func parseI18nTextToExcelFieldKey(text string) string {
	if v, ok := i18nMapByText[text]; ok {
		return v
	}
	return text
}

func decodeMapToIssueSheetModel(data *vars.DataForFulfill, m map[IssueSheetColumnUUID]excel.Column) (_ []vars.IssueSheetModel, err error) {
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
	models := make([]vars.IssueSheetModel, dataRowsNum, dataRowsNum)
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
			case fieldCommon:
				switch groupField {
				case fieldID:
					if cell.Value == "" { // update
						model.Common.ID = 0
						continue
					}
					id, err := strconv.ParseUint(cell.Value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("invalid id: %s", cell.Value)
					}
					model.Common.ID = id
				case fieldIterationName:
					model.Common.IterationName = cell.Value
				case fieldIssueType:
					issueType, err := parseStringIssueType(cell.Value)
					if err != nil {
						return nil, err
					}
					model.Common.IssueType = issueType
				case fieldIssueTitle:
					model.Common.IssueTitle = cell.Value
				case fieldContent:
					model.Common.Content = cell.Value
				case fieldState:
					model.Common.State = cell.Value
				case fieldPriority:
					priority, err := parseStringPriority(cell.Value)
					if err != nil {
						return nil, err
					}
					model.Common.Priority = priority
				case fieldComplexity:
					complexity, err := parseStringComplexity(cell.Value)
					if err != nil {
						return nil, err
					}
					model.Common.Complexity = complexity
				case fieldSeverity:
					severity, err := parseStringSeverity(cell.Value)
					if err != nil {
						return nil, err
					}
					model.Common.Severity = severity
				case fieldCreatorName:
					model.Common.CreatorName = cell.Value
				case fieldAssigneeName:
					model.Common.AssigneeName = cell.Value
				case fieldCreatedAt:
					model.Common.CreatedAt = mustParseStringTime(cell.Value, groupField)
				case fieldPlanStartedAt:
					model.Common.PlanStartedAt = mustParseStringTime(cell.Value, groupField)
				case fieldPlanFinishedAt:
					model.Common.PlanFinishedAt = mustParseStringTime(cell.Value, groupField)
				case fieldStartAt:
					model.Common.StartAt = mustParseStringTime(cell.Value, groupField)
				case fieldFinishAt:
					model.Common.FinishAt = mustParseStringTime(cell.Value, groupField)
				case fieldEstimateTime:
					model.Common.EstimateTime = cell.Value
				case fieldLabels:
					model.Common.Labels = vars.ParseStringSliceByComma(cell.Value)
				case fieldConnectionIssueIDs:
					var ids []int64
					for _, idStr := range vars.ParseStringSliceByComma(cell.Value) {
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
					// just skip
					continue
				}
			case fieldRequirementOnly:
				switch groupField {
				case fieldInclusionIssueIDs:
					var ids []int64
					for _, idStr := range vars.ParseStringSliceByComma(cell.Value) {
						id, err := parseStringIssueID(idStr)
						if err != nil {
							return nil, fmt.Errorf("invalid inclusion issue id: %s", idStr)
						}
						if id != nil {
							ids = append(ids, *id)
						}
					}
					model.RequirementOnly.InclusionIssueIDs = ids
				case fieldCustomFields:
					model.RequirementOnly.CustomFields = append(model.RequirementOnly.CustomFields, vars.ExcelCustomField{
						Title: parts[2],
						Value: cell.Value,
					})
				}
			case fieldTaskOnly:
				switch groupField {
				case fieldTaskType:
					model.TaskOnly.TaskType = cell.Value
				case fieldCustomFields:
					model.TaskOnly.CustomFields = append(model.TaskOnly.CustomFields, vars.ExcelCustomField{
						Title: parts[2],
						Value: cell.Value,
					})
				}
			case fieldBugOnly:
				switch groupField {
				case fieldOwnerName:
					model.BugOnly.OwnerName = cell.Value
				case fieldSource:
					model.BugOnly.Source = cell.Value
				case fieldReopenCount:
					if cell.Value == "" {
						model.BugOnly.ReopenCount = 0
						continue
					}
					reopenCount, err := strconv.ParseInt(cell.Value, 10, 32)
					if err != nil {
						return nil, fmt.Errorf("invalid reopen count: %s", cell.Value)
					}
					model.BugOnly.ReopenCount = int32(reopenCount)
				case fieldCustomFields:
					model.BugOnly.CustomFields = append(model.BugOnly.CustomFields, vars.ExcelCustomField{
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
	// replace '\' to ''
	s = strings.ReplaceAll(s, `\`, "")
	// parse as 2006-01-02 15:04:05
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err == nil {
		return &t, nil
	}
	// parse as 2023/10/9 0:00:00
	t, err = parseStringTimeSlash(s)
	if err == nil {
		return &t, nil
	}
	return nil, fmt.Errorf("invalid time: %s", s)
}

func parseStringTimeSlash(timeStr string) (time.Time, error) {
	// 定义时间字符串的格式
	layout := "2006/01/02 15:04:05"

	timeStr = strings.TrimSpace(timeStr)

	// 检查日期部分的位数
	parts := strings.SplitN(timeStr, " ", 2)
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid time: %s", timeStr)
	}

	dateParts := strings.SplitN(parts[0], "/", 3)
	if len(dateParts) < 3 {
		return time.Time{}, fmt.Errorf("invalid time: %s", timeStr)
	}
	if len(dateParts[1]) == 1 { // month
		dateParts[1] = "0" + dateParts[1]
	}
	if len(dateParts[2]) == 1 { // day
		dateParts[2] = "0" + dateParts[2]
	}
	timeStr = fmt.Sprintf(`%s/%s/%s %s`, dateParts[0], dateParts[1], dateParts[2], parts[1])

	// 解析时间字符串
	t, err := time.ParseInLocation(layout, timeStr, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
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
func CreateOrUpdateIssues(data *vars.DataForFulfill, issueSheetModels []vars.IssueSheetModel) (_ []*issuedao.Issue, issueModelMapByIssueID map[uint64]*vars.IssueSheetModel, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			fmt.Println(string(debug.Stack()))
		}
	}()
	idMapping := make(map[int64]uint64) // key: old id(can be negative for Excel Line Num, like L5), value: new id
	issueModelMapByIssueID = make(map[uint64]*vars.IssueSheetModel)
	var issues []*issuedao.Issue
	for i, model := range issueSheetModels {
		model := model
		// check state
		stateID, ok := data.StateMapByTypeAndName[model.Common.IssueType.String()][model.Common.State]
		if !ok {
			return nil, nil, fmt.Errorf("unknown state: %s, please contact project manager to add the corresponding status first", model.Common.State)
		}
		// check iteration
		iterationID := int64(-1) // default iteration value -1 for 待规划
		iteration, ok := data.IterationMapByName[model.Common.IterationName]
		if ok {
			iterationID = int64(iteration.ID)
		}
		if iterationID == 0 { // 0 is invalid, but iteration id is uint64, so lowest is 0
			iterationID = -1
		}
		issue := issuedao.Issue{
			BaseModel: dbengine.BaseModel{
				ID:        uint64(model.Common.ID),
				CreatedAt: vars.ChangePointerTimeToTime(model.Common.CreatedAt),
				UpdatedAt: vars.ChangePointerTimeToTime(model.Common.CreatedAt),
			},
			PlanStartedAt:  model.Common.PlanStartedAt,
			PlanFinishedAt: model.Common.PlanFinishedAt,
			ProjectID:      data.ProjectID,
			IterationID:    iterationID,
			AppID:          nil,
			RequirementID:  nil,
			Type:           model.Common.IssueType.String(),
			Title:          model.Common.IssueTitle,
			Content:        model.Common.Content,
			State:          stateID,
			Priority:       model.Common.Priority.String(),
			Complexity:     model.Common.Complexity.String(),
			Severity:       model.Common.Severity.String(),
			Creator:        data.ImportOnly.ProjectMemberIDByUserKey[model.Common.CreatorName],
			Assignee:       data.ImportOnly.ProjectMemberIDByUserKey[model.Common.AssigneeName],
			Source:         "",
			ManHour:        vars.MustGetJsonManHour(model.Common.EstimateTime),
			External:       true,
			Deleted:        false,
			Stage:          vars.GetIssueStage(model),
			Owner:          data.ImportOnly.ProjectMemberIDByUserKey[model.BugOnly.OwnerName],
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
		logrus.Debugf("issue-import: issue created or updated, id: %d", issue.ID)
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

func CreateIssueRelations(data *vars.DataForFulfill, issues []*issuedao.Issue, issueModelMapByIssueID map[uint64]*vars.IssueSheetModel) error {
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

// removeEmptySheetRows remove if row if all cells are empty
func removeEmptySheetRows(rows *[][]string) {
	var newRows [][]string
	for _, row := range *rows {
		var empty = true
		for _, cell := range row {
			if cell != "" {
				empty = false
				break
			}
		}
		if !empty {
			newRows = append(newRows, row)
		}
	}
	*rows = newRows
}

// autoFillEmptyRowCells fill empty cell if row is not empty
func autoFillEmptyRowCells(rows *[][]string, columnIndex int) {
	for i, row := range *rows {
		if len(row) < columnIndex {
			(*rows)[i] = append(row, make([]string, columnIndex-len(row))...)
		}
	}
}
