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
	"reflect"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets/sheet_customfield"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	streamcommon "github.com/erda-project/erda/internal/apps/dop/providers/issue/stream/common"
	"github.com/erda-project/erda/pkg/common/pbutil"
	"github.com/erda-project/erda/pkg/strutil"
)

func (h *Handler) ExportSheet(data *vars.DataForFulfill) (*sheets.RowsForExport, error) {
	mapByColumns, err := genIssueSheetTitleAndDataByColumn(data)
	if err != nil {
		return nil, fmt.Errorf("failed to gen sheet title and data by column, err: %v", err)
	}
	rowsForExport, err := mapByColumns.ConvertToExcelSheet(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to excel sheet, err: %v", err)
	}
	return rowsForExport, nil
}

// genIssueSheetTitleAndDataByColumn
func genIssueSheetTitleAndDataByColumn(data *vars.DataForFulfill) (*IssueSheetModelCellInfoByColumns, error) {
	models, err := getIssueSheetModels(data)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue sheet models, err: %v", err)
	}
	// 返回值
	info := NewIssueSheetModelCellInfoByColumns()
	// 反射
	for _, model := range models {
		modelValue := reflect.ValueOf(&model).Elem()
		for i := 0; i < modelValue.NumField(); i++ {
			groupField := modelValue.Field(i)
			groupStructField := modelValue.Type().Field(i)
			var uuid IssueSheetColumnUUID
			uuid.AddPart(getStructFieldExcelTag(groupStructField))
			// parse group field
			for j := 0; j < groupField.NumField(); j++ {
				valueField := groupField.Field(j)
				structField := groupField.Type().Field(j)
				uuid := uuid
				thisPart := getStructFieldExcelTag(structField)
				if thisPart == "-" {
					continue // ignore `-` field, like `lineNum`
				}
				uuid.AddPart(thisPart)
				// custom fields 动态字段，返回多个 column cell
				if structField.Type == reflect.TypeOf([]vars.ExcelCustomField{}) {
					// 遍历 customFields, 每个 cf 生成一个 column cell
					cfs := valueField.Interface().([]vars.ExcelCustomField)
					if len(cfs) == 0 {
						cfBelongingIssueType := getCustomFieldBelongingTypeFromUUID(uuid)
						cfs = vars.FormatIssueCustomFields(&pb.Issue{Id: int64(model.Common.ID)}, sheet_customfield.MustGetIssuePropertyEnumTypeByIssueType(cfBelongingIssueType), data)
					}
					for _, cf := range cfs {
						uuid := uuid
						uuid.AddPart(cf.Title)
						info.Add(uuid, strutil.String(cf.Value))
					}
				} else { // 其他字段，直接取值
					// set custom key for some fields for i18n
					cellValue := getStringCellValue(structField, valueField)
					cellValue = makeI18nKey(thisPart, cellValue)
					info.Add(uuid, cellValue)
				}
			}
		}
	}
	return &info, nil
}

func genDataValidationTip(data *vars.DataForFulfill, fieldName string, customFieldName string) (title, msg string) {
	switch fieldName {
	case fieldCustomFields:
		for _, concreteIssueTypeCfs := range data.CustomFieldMapByTypeName {
			for name, cf := range concreteIssueTypeCfs {
				if name == customFieldName {
					if cf.PropertyType == pb.PropertyTypeEnum_MultiSelect {
						title = data.I18n(makeI18nFieldTipTitleKey("MultiSelect"))
						var dp []string
						for _, ev := range cf.EnumeratedValues {
							dp = append(dp, ev.Name)
						}
						msg = strutil.Join(dp, "\n")
						return
					}
				}
			}
		}
	case fieldInclusionIssueIDs:
		title = data.I18n(makeI18nFieldTipTitleKey(fieldInclusionIssueIDs))
		msg = data.I18n(makeI18nFieldTipMsgKey(fieldInclusionIssueIDs))
	case fieldConnectionIssueIDs:
		title = data.I18n(makeI18nFieldTipTitleKey(fieldConnectionIssueIDs))
		msg = data.I18n(makeI18nFieldTipMsgKey(fieldConnectionIssueIDs))
	}
	return
}

func makeI18nFieldTipTitleKey(fieldKey string) string { return "tipTitleFor" + fieldKey }
func makeI18nFieldTipMsgKey(fieldKey string) string   { return "tipMsgFor" + fieldKey }

func genDropList(data *vars.DataForFulfill, fieldName string, customFieldName string) []string {
	switch fieldName {
	case fieldIterationName:
		return getIterationDropList(data)
	case fieldIssueType:
		return getFieldIssueTypeDropList(data)
	case FieldState:
		var dp []string
		for issueType, v := range data.StateMapByTypeAndName {
			if strutil.InSlice(issueType, supportedIssueTypes) {
				dp = append(dp, fmt.Sprintf("---%s---", data.I18n(makeI18nKey(fieldIssueType, issueType.String()))))
				for stateName := range v {
					dp = append(dp, stateName)
				}
			}
		}
		return dp
	case fieldPriority:
		return getFieldPriorityDropList(data)
	case fieldComplexity:
		return getFieldComplexityDropList(data)
	case fieldSeverity:
		return getFieldSeverityDropList(data)
	case fieldCreatorName, fieldAssigneeName, fieldOwnerName:
		return getUserRelatedDropList(data)
	case fieldTaskType:
		return getFieldTaskTypeDropList(data)
	case fieldSource:
		return getFieldSourceDropList(data)
	case fieldCustomFields:
		return getFieldCustomFieldsDropList(data, customFieldName)
	default:
		return nil
	}
}

func getCustomFieldBelongingTypeFromUUID(uuid IssueSheetColumnUUID) pb.IssueTypeEnum_Type {
	switch uuid.Decode()[0] {
	case fieldRequirementOnly:
		return pb.IssueTypeEnum_REQUIREMENT
	case fieldTaskOnly:
		return pb.IssueTypeEnum_TASK
	case fieldBugOnly:
		return pb.IssueTypeEnum_BUG
	default:
		panic(fmt.Errorf("failed to get issue type from uuid: %s", uuid.Decode()[0]))
	}
}

func getIssueSheetModels(data *vars.DataForFulfill) ([]vars.IssueSheetModel, error) {
	models := make([]vars.IssueSheetModel, 0, len(data.ExportOnly.Issues))
	if data.ExportOnly.IsDownloadTemplate {
		models = GenerateSampleIssueSheetModels(data)
		return models, nil
	}
	for _, issue := range data.ExportOnly.Issues {
		var model vars.IssueSheetModel
		// iteration name
		iterationName, err := getIterationName(data, issue)
		if err != nil {
			return nil, fmt.Errorf("failed to get iteration name, err: %v", err)
		}
		model.Common = vars.IssueSheetModelCommon{
			ID:                 uint64(issue.Id),
			IterationName:      iterationName,
			IssueType:          issue.Type,
			IssueTitle:         issue.Title,
			Content:            issue.Content,
			State:              data.StateMap[issue.State],
			Priority:           issue.Priority,
			Complexity:         issue.Complexity,
			Severity:           issue.Severity,
			CreatorName:        getUserNick(data, issue.Creator),
			AssigneeName:       getUserNick(data, issue.Assignee),
			CreatedAt:          pbutil.GetTimeInLocal(issue.CreatedAt),
			PlanStartedAt:      pbutil.GetTimeInLocal(issue.PlanStartedAt),
			PlanFinishedAt:     pbutil.GetTimeInLocal(issue.PlanFinishedAt),
			StartAt:            pbutil.GetTimeInLocal(issue.StartTime),
			FinishAt:           pbutil.GetTimeInLocal(issue.FinishTime),
			EstimateTime:       streamcommon.GetFormartTime(issue.IssueManHour, "EstimateTime"),
			Labels:             issue.Labels,
			ConnectionIssueIDs: data.ExportOnly.ConnectionMap[issue.Id],
		}
		model.RequirementOnly = vars.IssueSheetModelRequirementOnly{
			InclusionIssueIDs: data.ExportOnly.InclusionMap[issue.Id],
			CustomFields:      vars.FormatIssueCustomFields(issue, pb.PropertyIssueTypeEnum_REQUIREMENT, data),
		}
		model.TaskOnly = vars.IssueSheetModelTaskOnly{
			TaskType:     getIssueStageName(data, issue, pb.IssueTypeEnum_TASK),
			CustomFields: vars.FormatIssueCustomFields(issue, pb.PropertyIssueTypeEnum_TASK, data),
		}
		model.BugOnly = vars.IssueSheetModelBugOnly{
			OwnerName:    getUserNick(data, issue.Owner),
			Source:       getIssueStageName(data, issue, pb.IssueTypeEnum_BUG),
			ReopenCount:  issue.ReopenCount,
			CustomFields: vars.FormatIssueCustomFields(issue, pb.PropertyIssueTypeEnum_BUG, data),
		}
		models = append(models, model)
	}
	return models, nil
}

func getStructFieldExcelTag(structField reflect.StructField) string {
	tag := structField.Tag.Get("excel")
	if tag == "" {
		tag = structField.Name
	}
	return tag
}

func getStringCellValue(structField reflect.StructField, fieldValue reflect.Value) string {
	switch structField.Type {
	case reflect.TypeOf(&timestamppb.Timestamp{}):
		if fieldValue.IsNil() {
			return ""
		}
		t := pbutil.GetTimeInLocal(fieldValue.Interface().(*timestamppb.Timestamp))
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02 15:04:05")
	case reflect.TypeOf(&time.Time{}):
		if fieldValue.IsNil() {
			return ""
		}
		t := fieldValue.Interface().(*time.Time)
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02 15:04:05")
	case reflect.TypeOf([]int64{}): // ConnectionIssueIDs, InclusionIssueIDs
		ss := make([]string, 0, len(fieldValue.Interface().([]int64)))
		for _, i := range fieldValue.Interface().([]int64) {
			if i < 0 {
				ss = append(ss, fmt.Sprintf("L%d", -i))
			} else {
				ss = append(ss, strutil.String(i))
			}
		}
		return strings.Join(ss, ",")
	case reflect.TypeOf([]string{}):
		return strings.Join(fieldValue.Interface().([]string), ",")
	case reflect.TypeOf(int32(0)), reflect.TypeOf(uint64(0)):
		if fieldValue.IsZero() {
			return ""
		}
		fallthrough
	default:
		return strutil.String(fieldValue.Interface())
	}
}

func getIterationName(data *vars.DataForFulfill, issue *pb.Issue) (string, error) {
	iteration, ok := data.IterationMapByID[issue.IterationID]
	if !ok {
		return "", fmt.Errorf("iteration not found, issue id: %d, iteration id: %d", issue.Id, issue.IterationID)
	}
	return iteration.Title, nil
}

func getUserNick(data *vars.DataForFulfill, userid string) string {
	if userid == "" {
		return ""
	}
	if u, ok := data.ProjectMemberByUserID[userid]; ok {
		return u.Nick
	}
	return ""
}

func getIssueStageName(data *vars.DataForFulfill, issue *pb.Issue, targetIssueType pb.IssueTypeEnum_Type) string {
	if issue.Type != targetIssueType {
		return ""
	}
	stage := query.IssueStage{
		Type:  issue.Type.String(),
		Value: getIssueStageValue(issue),
	}
	v, ok := data.StageMap[stage]
	if ok {
		return v
	}
	return getIssueStageValue(issue)
}

func getIssueStageValue(issue *pb.Issue) string {
	switch issue.Type {
	case pb.IssueTypeEnum_TASK:
		return issue.TaskType
	case pb.IssueTypeEnum_BUG:
		return issue.BugStage
	default:
		return ""
	}
}
