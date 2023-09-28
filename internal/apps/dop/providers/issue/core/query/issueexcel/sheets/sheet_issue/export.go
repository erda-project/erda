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

	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	streamcommon "github.com/erda-project/erda/internal/apps/dop/providers/issue/stream/common"
	"github.com/erda-project/erda/pkg/common/pbutil"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
)

func (h *Handler) ExportSheet(data *vars.DataForFulfill) (excel.Rows, error) {
	mapByColumns, err := genIssueSheetTitleAndDataByColumn(data)
	if err != nil {
		return nil, fmt.Errorf("failed to gen sheet title and data by column, err: %v", err)
	}
	excelRows, err := mapByColumns.ConvertToExcelSheet()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to excel sheet, err: %v", err)
	}
	return excelRows, nil
}

func genIssueSheetTitleAndDataByColumn(data *vars.DataForFulfill) (*IssueSheetModelCellInfoByColumns, error) {
	models, err := getIssueSheetModels(data)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue sheet models, err: %v", err)
	}
	// 返回值
	info := IssueSheetModelCellInfoByColumns{
		M:            make(IssueSheetModelCellMapByColumns),
		OrderedUUIDs: make([]IssueSheetColumnUUID, 0),
	}
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
				uuid.AddPart(getStructFieldExcelTag(structField))
				// custom fields 动态字段，返回多个 column cell
				if structField.Type == reflect.TypeOf([]vars.ExcelCustomField{}) {
					// 遍历 customFields, 每个 cf 生成一个 column cell
					for _, cf := range valueField.Interface().([]vars.ExcelCustomField) {
						uuid := uuid
						uuid.AddPart(cf.Title)
						info.Add(uuid, strutil.String(cf.Value))
					}
				} else { // 其他字段，直接取值
					info.Add(uuid, getStringCellValue(structField, valueField))
				}
			}
		}
	}
	return &info, nil
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
			UpdatedAt:          pbutil.GetTimeInLocal(issue.UpdatedAt),
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
			TaskType:     issue.TaskType,
			CustomFields: vars.FormatIssueCustomFields(issue, pb.PropertyIssueTypeEnum_TASK, data),
		}
		model.BugOnly = vars.IssueSheetModelBugOnly{
			OwnerName:    getUserNick(data, issue.Owner),
			Source:       issue.BugStage,
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
	case reflect.TypeOf(&timestamp.Timestamp{}):
		if fieldValue.IsNil() {
			return ""
		}
		t := pbutil.GetTimeInLocal(fieldValue.Interface().(*timestamp.Timestamp))
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
