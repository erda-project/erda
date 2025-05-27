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
	"github.com/tealeg/xlsx/v3"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets/sheet_customfield"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	issuedao "github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
)

type Handler struct{ sheets.DefaultImporter }

func (h *Handler) SheetName() string { return vars.NameOfSheetIssue }

func (h *Handler) DecodeSheet(data *vars.DataForFulfill, s *excel.Sheet) error {
	InitI18nMap(data)
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

	info := NewIssueSheetModelCellInfoByColumns()
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
			info.Add(uuid, row[i])
		}
	}
	// decode map to models
	// 只校验数据格式，不做关联判断，关联判断由具体 sheet handler 处理
	models, err := decodeMapToIssueSheetModel(data, info)
	if err != nil {
		return fmt.Errorf("failed to decode issue sheet, err: %v", err)
	}
	data.ImportOnly.Sheets.Must.IssueInfo = models
	return nil
}

func (h *Handler) BeforeCreateIssues(data *vars.DataForFulfill) error {
	return nil
}

func (h *Handler) CreateIssues(data *vars.DataForFulfill) error {
	issues, issueModelMapByIssueID, err := createOrUpdateIssues(data, data.ImportOnly.Sheets.Must.IssueInfo)
	if err != nil {
		return fmt.Errorf("failed to create or update issues, err: %v", err)
	}
	data.ImportOnly.Created = vars.Created{
		Issues:                 issues,
		IssueModelMapByIssueID: issueModelMapByIssueID,
	}
	return nil
}

func (h *Handler) AfterCreateIssues(data *vars.DataForFulfill) error {
	if err := CreateIssueRelations(data, data.ImportOnly.Created.Issues, data.ImportOnly.Created.IssueModelMapByIssueID); err != nil {
		return fmt.Errorf("failed to create issue relations, err: %v", err)
	}
	return nil
}

func (h *Handler) AppendErrorColumn(data *vars.DataForFulfill, sheet *excel.Sheet) {
	// check collected errors
	if len(data.ImportOnly.Errs) == 0 {
		return
	}
	currentLineNum := 1
	totalColumnNum := len(sheet.UnmergedSlice[0]) + 1
	haveErrorColumn := false
	errorColumnIndex := -1
	_ = sheet.XlsxSheet.ForEachRow(func(r *xlsx.Row) error {
		defer func() {
			currentLineNum++
		}()
		// check if error column already exists
		currentColumnIndex := 0
		if currentLineNum == 1 {
			_ = r.ForEachCell(func(c *xlsx.Cell) error {
				defer func() {
					currentColumnIndex++
				}()
				if strutil.InSlice(c.Value, data.AllI18nValuesByKey(fieldError)) {
					haveErrorColumn = true
					errorColumnIndex = currentColumnIndex
					return nil
				}
				return nil
			})
		}
		// if error column have, just update it; or we need add cell
		var cell *xlsx.Cell
		if haveErrorColumn {
			cell = r.GetCell(errorColumnIndex)
		} else {
			cell = r.AddCell()
		}
		cell.Value = "" // reset before assign
		if currentLineNum <= 3 {
			if currentLineNum == 1 {
				cell.VMerge = 2
				cell.Value = data.I18n(fieldError)
				style := excel.DefaultTitleCellStyle()
				style.Fill.FgColor = "FFFF0000"
				cell.SetStyle(style)
			}
			return nil
		}
		for _, err := range data.ImportOnly.Errs {
			if err.LineNum == currentLineNum {
				if cell.Value == "" {
					cell.Value = err.Msg
				} else {
					cell.Value += "; " + err.Msg
				}
			}
		}
		return nil
	})
	widthHandler := excel.NewSheetHandlerForAutoColWidth(totalColumnNum)
	_ = widthHandler(sheet.XlsxSheet)
}

func GetI18nErr(data *vars.DataForFulfill, fieldKey string, args ...interface{}) string {
	return getI18nErr2(data, []string{fieldKey}, args...)
}

func getI18nErr2(data *vars.DataForFulfill, fieldKeys []string, args ...interface{}) string {
	v := data.I18n("invalid")
	for _, fieldKey := range fieldKeys {
		v += data.I18n(fieldKey)
	}
	if len(args) > 0 {
		v = fmt.Sprintf("%s: %v", v, args)
	}
	return v
}

func parseI18nTextToExcelFieldKey(text string) string {
	if v, ok := i18nMapByText[text]; ok {
		return v
	}
	return text
}

func getAvailableIssueIDs(data *vars.DataForFulfill, idColumn excel.Column) {
	for i, cell := range idColumn {
		// add line num
		lineNum := i + (uuidPartsMustLength + 1)
		data.ImportOnly.AvailableIssueIDsMap[-int64(lineNum)] = 0
		// add cell value if have
		if cell.Value != "" {
			id, err := strconv.ParseUint(cell.Value, 10, 64)
			if err != nil {
				data.AppendImportError(lineNum, GetI18nErr(data, fieldID, cell.Value))
				continue
			}
			data.ImportOnly.AvailableIssueIDsMap[int64(id)] = 0
		}
	}
	// add current project issue ids
	for id := range data.ImportOnly.CurrentProjectIssueMap {
		data.ImportOnly.AvailableIssueIDsMap[int64(id)] = id
	}
}

// decodeMapToIssueSheetModel
// check cell value when decode
func decodeMapToIssueSheetModel(data *vars.DataForFulfill, info IssueSheetModelCellInfoByColumns) (_ []vars.IssueSheetModel, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			// print runtime stacktrace
			fmt.Println(string(debug.Stack()))
		}
	}()
	// prepare models
	m := info.M
	var dataRowsNum int
	for _, column := range m {
		dataRowsNum = len(column)
		break
	}
	// get all ids for id check use
	getAvailableIssueIDs(data, m[NewIssueSheetColumnUUID(fieldCommon, fieldID, fieldID)])
	if err != nil {
		return nil, fmt.Errorf("failed to get available issue ids, err: %v", err)
	}

	models := make([]vars.IssueSheetModel, dataRowsNum, dataRowsNum)
	for _, uuid := range info.OrderedUUIDs {
		column := m[uuid]
		parts := uuid.Decode()
		if len(parts) != uuidPartsMustLength {
			return nil, fmt.Errorf("invalid uuid: %s", uuid)
		}
		for i, cell := range column {
			model := &models[i]
			model.Common.LineNum = i + (uuidPartsMustLength + 1) // Excel line num from 1
			groupType := parts[0]
			groupField := parts[1]
			i18nErrV := GetI18nErr(data, groupField, cell.Value)
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
						data.AppendImportError(model.Common.LineNum, i18nErrV)
						continue
					}
					model.Common.ID = id
				case fieldIterationName:
					model.Common.IterationName = cell.Value
				case fieldIssueType:
					issueType, err := parseStringIssueType(data, cell.Value)
					if err != nil {
						data.AppendImportError(model.Common.LineNum, i18nErrV)
						continue
					}
					model.Common.IssueType = *issueType
				case fieldIssueTitle:
					model.Common.IssueTitle = cell.Value
				case fieldContent:
					model.Common.Content = cell.Value
				case FieldState:
					model.Common.State = cell.Value
				case fieldPriority:
					priority, err := parseStringPriority(data, cell.Value)
					if err != nil {
						data.AppendImportError(model.Common.LineNum, i18nErrV)
						continue
					}
					model.Common.Priority = *priority
				case fieldComplexity:
					complexity, err := parseStringComplexity(data, cell.Value)
					if err != nil {
						data.AppendImportError(model.Common.LineNum, i18nErrV)
						continue
					}
					model.Common.Complexity = *complexity
				case fieldSeverity:
					severity, err := parseStringSeverity(data, cell.Value)
					if err != nil {
						data.AppendImportError(model.Common.LineNum, i18nErrV)
						continue
					}
					model.Common.Severity = *severity
				case fieldCreatorName:
					model.Common.CreatorName = cell.Value
				case fieldAssigneeName:
					model.Common.AssigneeName = cell.Value
				case fieldCreatedAt:
					createdAt, err := parseStringTime(cell.Value)
					if err != nil {
						data.AppendImportError(model.Common.LineNum, i18nErrV)
						continue
					}
					model.Common.CreatedAt = createdAt
				case fieldPlanStartedAt:
					planStartedAt, err := parseStringTime(cell.Value)
					if err != nil {
						data.AppendImportError(model.Common.LineNum, i18nErrV)
						continue
					}
					model.Common.PlanStartedAt = planStartedAt
				case fieldPlanFinishedAt:
					planFinishedAt, err := parseStringTime(cell.Value)
					if err != nil {
						data.AppendImportError(model.Common.LineNum, i18nErrV)
						continue
					}
					model.Common.PlanFinishedAt = planFinishedAt
				case fieldStartAt:
					startedAt, err := parseStringTime(cell.Value)
					if err != nil {
						data.AppendImportError(model.Common.LineNum, i18nErrV)
						continue
					}
					model.Common.StartAt = startedAt
				case fieldFinishAt:
					finishedAt, err := parseStringTime(cell.Value)
					if err != nil {
						data.AppendImportError(model.Common.LineNum, i18nErrV)
						continue
					}
					model.Common.FinishAt = finishedAt
				case fieldEstimateTime:
					if _, err := vars.NewManhour(cell.Value); err != nil {
						data.AppendImportError(model.Common.LineNum, i18nErrV)
						continue
					}
					model.Common.EstimateTime = cell.Value
				case fieldLabels:
					model.Common.Labels = vars.ParseStringSliceByComma(cell.Value)
				case fieldConnectionIssueIDs:
					var ids []int64
					for _, idStr := range vars.ParseStringSliceByComma(cell.Value) {
						id, err := parseStringIssueID(idStr)
						if err != nil {
							data.AppendImportError(model.Common.LineNum, i18nErrV)
							continue
						}
						if id != nil {
							// check id is available
							if _, ok := data.ImportOnly.AvailableIssueIDsMap[*id]; !ok {
								data.AppendImportError(model.Common.LineNum, GetI18nErr(data, fieldConnectionIssueIDs, idStr))
								continue
							}
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
							data.AppendImportError(model.Common.LineNum, GetI18nErr(data, fieldInclusionIssueIDs, idStr))
							continue
						}
						if id != nil {
							// check id is available
							if _, ok := data.ImportOnly.AvailableIssueIDsMap[*id]; !ok {
								data.AppendImportError(model.Common.LineNum, GetI18nErr(data, fieldInclusionIssueIDs, idStr))
								continue
							}
							ids = append(ids, *id)
						}
					}
					model.RequirementOnly.InclusionIssueIDs = ids
				case fieldCustomFields:
					parseCustomField(data, pb.IssueTypeEnum_REQUIREMENT, model, parts[2], cell.Value, &model.RequirementOnly.CustomFields)
				}
			case fieldTaskOnly:
				switch groupField {
				case fieldTaskType:
					taskType, err := parseStringTaskType(data, cell.Value)
					if err != nil {
						data.AppendImportError(model.Common.LineNum, i18nErrV)
						continue
					}
					model.TaskOnly.TaskType = taskType
				case fieldCustomFields:
					parseCustomField(data, pb.IssueTypeEnum_TASK, model, parts[2], cell.Value, &model.TaskOnly.CustomFields)
				}
			case fieldBugOnly:
				switch groupField {
				case fieldOwnerName:
					model.BugOnly.OwnerName = cell.Value
				case fieldSource:
					source, err := parseStringSource(data, cell.Value)
					if err != nil {
						data.AppendImportError(model.Common.LineNum, getI18nErr2(data,
							[]string{
								makeI18nKey(i18nKeyPrefixOfIssueType, model.Common.IssueType.String()),
								fieldCustomFields,
								parts[2],
							},
							cell.Value))
						continue
					}
					model.BugOnly.Source = source
				case fieldReopenCount:
					if cell.Value == "" {
						model.BugOnly.ReopenCount = 0
						continue
					}
					reopenCount, err := strconv.ParseInt(cell.Value, 10, 32)
					if err != nil {
						data.AppendImportError(model.Common.LineNum, i18nErrV)
						continue
					}
					model.BugOnly.ReopenCount = int32(reopenCount)
				case fieldCustomFields:
					parseCustomField(data, pb.IssueTypeEnum_BUG, model, parts[2], cell.Value, &model.BugOnly.CustomFields)
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

// createOrUpdateIssues 创建或更新 issues
// 根据 project id 进行判断
// 更新 model 里的相关关联 ID 字段，比如 L1 转换为具体的 ID
func createOrUpdateIssues(data *vars.DataForFulfill, issueSheetModels []vars.IssueSheetModel) (_ []*issuedao.Issue, issueModelMapByIssueID map[uint64]*vars.IssueSheetModel, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			fmt.Println(string(debug.Stack()))
		}
	}()
	issueModelMapByIssueID = make(map[uint64]*vars.IssueSheetModel)
	var issues []*issuedao.Issue
	for _, model := range issueSheetModels {
		model := model
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
			State:          data.StateMapByTypeAndName[model.Common.IssueType][model.Common.State],
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
		if model.Common.ID > 0 {
			data.ImportOnly.AvailableIssueIDsMap[int64(model.Common.ID)] = issue.ID
		}
		data.ImportOnly.AvailableIssueIDsMap[-int64(model.Common.LineNum)] = issue.ID // excel line num, star from 1
	}
	// all id mapping done here, update old ids
	for i := range issueSheetModels {
		model := issueSheetModels[i]
		// id
		model.Common.ID = data.ImportOnly.AvailableIssueIDsMap[int64(model.Common.ID)]
		// connection issue id
		for j := range model.Common.ConnectionIssueIDs {
			newID, ok := data.ImportOnly.AvailableIssueIDsMap[model.Common.ConnectionIssueIDs[j]]
			if !ok {
				return nil, nil, fmt.Errorf("invalid connection issue id: %d", model.Common.ConnectionIssueIDs[j])
			}
			model.Common.ConnectionIssueIDs[j] = int64(newID)
		}
		// inclusion issue id
		for j := range model.RequirementOnly.InclusionIssueIDs {
			newID, ok := data.ImportOnly.AvailableIssueIDsMap[model.RequirementOnly.InclusionIssueIDs[j]]
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

var supportedIssueTypes = []pb.IssueTypeEnum_Type{pb.IssueTypeEnum_REQUIREMENT, pb.IssueTypeEnum_TASK, pb.IssueTypeEnum_BUG, pb.IssueTypeEnum_TICKET}
var (
	i18nKeyPrefixOfIssueType  = "issue_type:"
	i18nKeyPrefixOfPriority   = "priority:"
	i18nKeyPrefixOfComplexity = "complexity:"
	i18nKeyPrefixOfSeverity   = "severity:"

	i18nKeyPrefixMapByFieldKey = map[string]string{
		fieldIssueType:  i18nKeyPrefixOfIssueType,
		fieldPriority:   i18nKeyPrefixOfPriority,
		fieldComplexity: i18nKeyPrefixOfComplexity,
		fieldSeverity:   i18nKeyPrefixOfSeverity,
	}
)

func getDataCellI18nValue(data *vars.DataForFulfill, fieldKey string, cellValue string) string {
	if cellValue == "" {
		return ""
	}
	if _, ok := i18nKeyPrefixMapByFieldKey[fieldKey]; !ok {
		return cellValue
	}
	return data.I18n(cellValue)
}

func makeI18nKey(fieldKey string, args ...string) string {
	if len(args) == 0 {
		return fieldKey
	}
	arg := args[0]
	prefix, ok := i18nKeyPrefixMapByFieldKey[fieldKey]
	if !ok {
		return arg
	}
	return prefix + arg
}

func getFieldIssueTypeDropList(data *vars.DataForFulfill) []string {
	var dp []string
	for _, issueType := range supportedIssueTypes {
		s := data.I18n(i18nKeyPrefixOfIssueType + issueType.String())
		dp = append(dp, s)
	}
	return dp
}

func getFieldPriorityDropList(data *vars.DataForFulfill) []string {
	var dp []string
	for i := 0; i < len(pb.IssuePriorityEnum_Priority_name); i++ {
		priority := pb.IssuePriorityEnum_Priority_name[int32(i)]
		s := data.I18n(makeI18nKey(fieldPriority, priority))
		dp = append(dp, s)
	}
	return dp
}

func getFieldComplexityDropList(data *vars.DataForFulfill) []string {
	var dp []string
	for i := 0; i < len(pb.IssueComplexityEnum_Complextity_name); i++ {
		complexity := pb.IssueComplexityEnum_Complextity_name[int32(i)]
		s := data.I18n(makeI18nKey(fieldComplexity, complexity))
		dp = append(dp, s)
	}
	return dp
}

func getFieldSeverityDropList(data *vars.DataForFulfill) []string {
	var dp []string
	for i := 0; i < len(pb.IssueSeverityEnum_Severity_name); i++ {
		severity := pb.IssueSeverityEnum_Severity_name[int32(i)]
		s := data.I18n(makeI18nKey(fieldSeverity, severity))
		dp = append(dp, s)
	}
	return dp
}

func getFieldTaskTypeDropList(data *vars.DataForFulfill) []string {
	var dp []string
	for kv, name := range data.StageMap {
		if kv.Type == pb.IssueTypeEnum_TASK.String() {
			dp = append(dp, name)
		}
	}
	return dp
}

func getFieldSourceDropList(data *vars.DataForFulfill) []string {
	var dp []string
	for kv, name := range data.StageMap {
		if kv.Type == pb.IssueTypeEnum_BUG.String() {
			dp = append(dp, name)
		}
	}
	return dp
}

func getIterationDropList(data *vars.DataForFulfill) []string {
	var dp []string
	for _, iteration := range data.IterationMapByID {
		dp = append(dp, iteration.Title)
	}
	return dp
}

func getFieldCustomFieldsDropList(data *vars.DataForFulfill, customFieldName string) []string {
	for _, concreteIssueTypeCfs := range data.CustomFieldMapByTypeName {
		for name, cf := range concreteIssueTypeCfs {
			if name == customFieldName {
				if cf.PropertyType == pb.PropertyTypeEnum_Person {
					return getUserRelatedDropList(data)
				}
				if cf.PropertyType == pb.PropertyTypeEnum_Select {
					var dp []string
					for _, ev := range cf.EnumeratedValues {
						dp = append(dp, ev.Name)
					}
					return dp
				}
				return nil
			}
		}
	}
	return nil
}

func getUserRelatedDropList(data *vars.DataForFulfill) []string {
	var dp []string
	for _, member := range data.ProjectMemberByUserID {
		dp = append(dp, member.Nick)
	}
	return strutil.DedupSlice(dp, true)
}

func parseStringIssueType(data *vars.DataForFulfill, input string) (*pb.IssueTypeEnum_Type, error) {
	kvs := []struct {
		i18nKeys    []string
		matchedType pb.IssueTypeEnum_Type
	}{
		{
			i18nKeys:    append(data.AllI18nValuesByKey(makeI18nKey(fieldIssueType, pb.IssueTypeEnum_REQUIREMENT.String())), pb.IssueTypeEnum_REQUIREMENT.String()),
			matchedType: pb.IssueTypeEnum_REQUIREMENT,
		},
		{
			i18nKeys:    append(data.AllI18nValuesByKey(i18nKeyPrefixOfIssueType+pb.IssueTypeEnum_TASK.String()), pb.IssueTypeEnum_TASK.String()),
			matchedType: pb.IssueTypeEnum_TASK,
		},
		{
			i18nKeys:    append(data.AllI18nValuesByKey(i18nKeyPrefixOfIssueType+pb.IssueTypeEnum_BUG.String()), pb.IssueTypeEnum_BUG.String()),
			matchedType: pb.IssueTypeEnum_BUG,
		},
		{
			i18nKeys:    append(data.AllI18nValuesByKey(i18nKeyPrefixOfIssueType+pb.IssueTypeEnum_TICKET.String()), pb.IssueTypeEnum_TICKET.String()),
			matchedType: pb.IssueTypeEnum_TICKET,
		},
	}
	for _, kv := range kvs {
		if strutil.InSlice(input, kv.i18nKeys) {
			return &kv.matchedType, nil
		}
	}
	return nil, fmt.Errorf("invalid issue type: %s", input)
}

func parseStringPriority(data *vars.DataForFulfill, input string) (*pb.IssuePriorityEnum_Priority, error) {
	kvs := []struct {
		i18nKeys    []string
		matchedType pb.IssuePriorityEnum_Priority
	}{
		{
			i18nKeys:    data.AllI18nValuesByKey(i18nKeyPrefixOfPriority + pb.IssuePriorityEnum_LOW.String()),
			matchedType: pb.IssuePriorityEnum_LOW,
		},
		{
			i18nKeys:    data.AllI18nValuesByKey(i18nKeyPrefixOfPriority + pb.IssuePriorityEnum_NORMAL.String()),
			matchedType: pb.IssuePriorityEnum_NORMAL,
		},
		{
			i18nKeys:    data.AllI18nValuesByKey(i18nKeyPrefixOfPriority + pb.IssuePriorityEnum_HIGH.String()),
			matchedType: pb.IssuePriorityEnum_HIGH,
		},
		{
			i18nKeys:    data.AllI18nValuesByKey(i18nKeyPrefixOfPriority + pb.IssuePriorityEnum_URGENT.String()),
			matchedType: pb.IssuePriorityEnum_URGENT,
		},
	}
	for _, kv := range kvs {
		if strutil.InSlice(input, kv.i18nKeys) {
			return &kv.matchedType, nil
		}
	}
	// use default priority
	defaultPriority := pb.IssuePriorityEnum_NORMAL
	return &defaultPriority, nil
}

func parseStringComplexity(data *vars.DataForFulfill, input string) (*pb.IssueComplexityEnum_Complextity, error) {
	kvs := []struct {
		i18nKeys    []string
		matchedType pb.IssueComplexityEnum_Complextity
	}{
		{
			i18nKeys:    data.AllI18nValuesByKey(i18nKeyPrefixOfComplexity + pb.IssueComplexityEnum_EASY.String()),
			matchedType: pb.IssueComplexityEnum_EASY,
		},
		{
			i18nKeys:    data.AllI18nValuesByKey(i18nKeyPrefixOfComplexity + pb.IssueComplexityEnum_NORMAL.String()),
			matchedType: pb.IssueComplexityEnum_NORMAL,
		},
		{
			i18nKeys:    data.AllI18nValuesByKey(i18nKeyPrefixOfComplexity + pb.IssueComplexityEnum_HARD.String()),
			matchedType: pb.IssueComplexityEnum_HARD,
		},
	}
	for _, kv := range kvs {
		if strutil.InSlice(input, kv.i18nKeys) {
			return &kv.matchedType, nil
		}
	}
	// use default complexity
	defaultComplexity := pb.IssueComplexityEnum_NORMAL
	return &defaultComplexity, nil
}

func parseStringSeverity(data *vars.DataForFulfill, input string) (*pb.IssueSeverityEnum_Severity, error) {
	kvs := []struct {
		i18nKeys    []string
		matchedType pb.IssueSeverityEnum_Severity
	}{
		{
			i18nKeys:    data.AllI18nValuesByKey(i18nKeyPrefixOfSeverity + pb.IssueSeverityEnum_SUGGEST.String()),
			matchedType: pb.IssueSeverityEnum_SUGGEST,
		},
		{
			i18nKeys:    data.AllI18nValuesByKey(i18nKeyPrefixOfSeverity + pb.IssueSeverityEnum_SLIGHT.String()),
			matchedType: pb.IssueSeverityEnum_SLIGHT,
		},
		{
			i18nKeys:    data.AllI18nValuesByKey(i18nKeyPrefixOfSeverity + pb.IssueSeverityEnum_NORMAL.String()),
			matchedType: pb.IssueSeverityEnum_NORMAL,
		},
		{
			i18nKeys:    data.AllI18nValuesByKey(i18nKeyPrefixOfSeverity + pb.IssueSeverityEnum_SERIOUS.String()),
			matchedType: pb.IssueSeverityEnum_SERIOUS,
		},
		{
			i18nKeys:    data.AllI18nValuesByKey(i18nKeyPrefixOfSeverity + pb.IssueSeverityEnum_FATAL.String()),
			matchedType: pb.IssueSeverityEnum_FATAL,
		},
	}
	for _, kv := range kvs {
		if strutil.InSlice(input, kv.i18nKeys) {
			return &kv.matchedType, nil
		}
	}
	// use default severity
	defaultSeverity := pb.IssueSeverityEnum_NORMAL
	return &defaultSeverity, nil
}

func parseStringTaskType(data *vars.DataForFulfill, input string) (string, error) {
	if input == "" {
		return "", nil
	}
	for kv, name := range data.StageMap {
		if kv.Type == pb.IssueTypeEnum_TASK.String() && (name == input || kv.Value == input) {
			return kv.Value, nil
		}
	}
	// empty if not found
	return "", nil
}

func parseStringSource(data *vars.DataForFulfill, input string) (string, error) {
	if input == "" {
		return "", nil
	}
	for kv, name := range data.StageMap {
		if kv.Type == pb.IssueTypeEnum_BUG.String() && name == input {
			return kv.Value, nil
		}
	}
	// empty if not found
	return "", nil
}

func parseCustomField(data *vars.DataForFulfill, expectIssueType pb.IssueTypeEnum_Type, model *vars.IssueSheetModel, cfName, cfValue string, appendable *[]vars.ExcelCustomField) {
	if cfValue == "" {
		return
	}
	if model.Common.IssueType != expectIssueType { // skip check if issue type mismatch
		return
	}
	cf := vars.ExcelCustomField{
		Title: cfName,
		Value: cfValue,
	}
	if err := sheet_customfield.CheckCustomFieldValue(data, model.Common.IssueType, cf.Title, cf.Value); err != nil {
		data.AppendImportError(model.Common.LineNum, getI18nErr2(data,
			[]string{
				makeI18nKey(i18nKeyPrefixOfIssueType, model.Common.IssueType.String()),
				fieldCustomFields,
				"-",
				cfName,
			},
			cfValue))
		return
	}
	*appendable = append(*appendable, cf)
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
