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
	"reflect"
	"strings"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	streamcommon "github.com/erda-project/erda/internal/apps/dop/providers/issue/stream/common"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
)

type IssueSheetColumnUUID string
type IssueSheetColumnUUIDParts []string

const issueSheetColumnUUIDSplitter = "---"
const uuidPartsMustLength = 3

func NewIssueSheetColumnUUID(parts ...string) IssueSheetColumnUUID {
	return IssueSheetColumnUUID(strings.Join(parts, issueSheetColumnUUIDSplitter))
}
func (uuid *IssueSheetColumnUUID) Decode() IssueSheetColumnUUIDParts {
	if *uuid == "" {
		return nil
	}
	return strings.Split(string(*uuid), issueSheetColumnUUIDSplitter)
}
func (uuid *IssueSheetColumnUUID) AutoComplete() {
	parts := uuid.Decode()
	if len(parts) > uuidPartsMustLength {
		panic("issue sheet column uuid have max 3 parts")
	}
	if len(parts) == 0 {
		panic("issue sheet column uuid must have at least 1 part")
	}
	// auto fulfill with last elem if not enough to uuidPartsMustLength
	for len(parts) < uuidPartsMustLength {
		parts = append(parts, parts[len(parts)-1])
	}
	// set back to uuid
	*uuid = IssueSheetColumnUUID(strings.Join(parts, issueSheetColumnUUIDSplitter))
}
func (uuid *IssueSheetColumnUUID) String() string {
	uuid.AutoComplete()
	return string(*uuid)
}
func (uuid *IssueSheetColumnUUID) AddPart(part string) {
	parts := uuid.Decode()
	parts = append(parts, part)
	*uuid = IssueSheetColumnUUID(strings.Join(parts, issueSheetColumnUUIDSplitter))
}

type IssueSheetModelCellInfoByColumns struct {
	M            IssueSheetModelCellMapByColumns
	OrderedUUIDs []IssueSheetColumnUUID
}

type IssueSheetModelCellMapByColumns map[IssueSheetColumnUUID]excel.Column

// Add
// isDemoModel just used to set all custom fields' uuids in order, can't be added to truly data M
func (info *IssueSheetModelCellInfoByColumns) Add(uuid IssueSheetColumnUUID, cellValue string) {
	uuid.AutoComplete()
	// ordered uuids only add once
	if _, ok := info.M[uuid]; !ok {
		info.OrderedUUIDs = append(info.OrderedUUIDs, uuid)
	}
	info.M[uuid] = append(info.M[uuid], excel.Cell{Value: cellValue})
}

func (info *IssueSheetModelCellInfoByColumns) ConvertToExcelSheet() (excel.Rows, error) {
	// create [][]excel.Cell
	var dataRowLength int
	// get data row length
	for _, columnCells := range info.M {
		dataRowLength = len(columnCells)
		break
	}
	rows := make(excel.Rows, uuidPartsMustLength+dataRowLength)
	for i := range rows {
		rows[i] = make(excel.Row, len(info.M))
	}
	// set by (x,y)
	columnIndex := 0
	for _, uuid := range info.OrderedUUIDs {
		column, ok := info.M[uuid]
		if !ok {
			panic(fmt.Sprintf("uuid: %s not found in info.M", uuid))
		}
		// set column title cells
		parts := uuid.Decode()
		for i, uuidPart := range parts {
			rows[i][columnIndex] = excel.Cell{Value: uuidPart, Style: &excel.CellStyle{IsTitle: true}}
		}
		// auto merge title cells with same value
		autoMergeTitleCellsWithSameValue(rows[:uuidPartsMustLength])
		// set column data cells
		for i, cell := range column {
			rows[uuidPartsMustLength+i][columnIndex] = cell
		}
		columnIndex++
	}
	return rows, nil
}

// excel sheet 聚合时，主动对 title cell 进行探测
// 水平探测:
// - 从右向左探测，如果左侧的 cell 和当前 cell 一致，则左边 cell 的 HMergeNum+1
// - 依次向左
//
// 纵向探测:
// - 从下往上探测，如果上面的 cell 和当前 cell 一致，则上面的 cell 的 VMergeNum+1
func autoMergeTitleCellsWithSameValue(rows excel.Rows) {
	// horizontal merge
	for rowIndex := range rows {
		for columnIndex := len(rows[rowIndex]) - 1; columnIndex >= 0; columnIndex-- {
			if columnIndex == 0 {
				continue
			}
			if rows[rowIndex][columnIndex].Value == rows[rowIndex][columnIndex-1].Value {
				cell := rows[rowIndex][columnIndex]
				cell.HorizontalMergeNum++
				rows[rowIndex][columnIndex-1] = cell
			}
		}
	}
	// vertical merge
	for columnIndex := range rows[0] {
		for rowIndex := len(rows) - 1; rowIndex >= 0; rowIndex-- {
			if rowIndex == 0 {
				continue
			}
			if rows[rowIndex][columnIndex].Value == rows[rowIndex-1][columnIndex].Value {
				cell := rows[rowIndex][columnIndex]
				cell.VerticalMergeNum++
				rows[rowIndex-1][columnIndex] = cell
			}
		}
	}
}

func (data DataForFulfill) genIssueSheetTitleAndDataByColumn() (*IssueSheetModelCellInfoByColumns, error) {
	models, err := data.getIssueSheetModels()
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
			groupTypeField := modelValue.Type().Field(i)
			var uuid IssueSheetColumnUUID
			uuid.AddPart(getStructFieldExcelTag(groupTypeField))
			// parse group field
			for j := 0; j < groupField.NumField(); j++ {
				valueField := groupField.Field(j)
				typeField := groupField.Type().Field(j)
				uuid := uuid
				uuid.AddPart(getStructFieldExcelTag(typeField))
				// custom fields 动态字段，返回多个 column cell
				if typeField.Type == reflect.TypeOf([]ExcelCustomField{}) {
					// 遍历 customFields, 每个 cf 生成一个 column cell
					for _, cf := range valueField.Interface().([]ExcelCustomField) {
						uuid := uuid
						uuid.AddPart(cf.Title)
						info.Add(uuid, strutil.String(cf.Value))
					}
				} else { // 其他字段，直接取值
					info.Add(uuid, strutil.String(valueField.Interface()))
				}
			}
		}
	}
	return &info, nil
}

func (data DataForFulfill) getIssueSheetModels() ([]IssueSheetModel, error) {
	models := make([]IssueSheetModel, 0, len(data.Issues))
	for _, issue := range data.Issues {
		var model IssueSheetModel
		model.Common = IssueSheetModelCommon{
			ID:                 issue.Id,
			IterationName:      data.IterationMap[issue.IterationID],
			IssueType:          issue.Type,
			IssueTitle:         issue.Title,
			Content:            issue.Content,
			State:              data.StateMap[issue.State],
			Priority:           issue.Priority,
			Complexity:         issue.Complexity,
			Severity:           issue.Severity,
			CreatorName:        data.getUserNick(issue.Creator),
			AssigneeName:       data.getUserNick(issue.Assignee),
			CreatedAt:          formatTimeFromTimestamp(issue.CreatedAt),
			PlanStartedAt:      formatTimeFromTimestamp(issue.PlanStartedAt),
			PlanFinishedAt:     formatTimeFromTimestamp(issue.PlanFinishedAt),
			StartAt:            formatTimeFromTimestamp(issue.StartTime),
			FinishAt:           formatTimeFromTimestamp(issue.FinishTime),
			EstimateTime:       streamcommon.GetFormartTime(issue.IssueManHour, "EstimateTime"),
			Labels:             issue.Labels,
			ConnectionIssueIDs: data.ConnectionMap[issue.Id],
		}
		model.RequirementOnly = IssueSheetModelRequirementOnly{
			InclusionIssueIDs: data.InclusionMap[issue.Id],
			CustomFields:      formatIssueCustomFields(issue, pb.PropertyIssueTypeEnum_REQUIREMENT, data),
		}
		model.TaskOnly = IssueSheetModelTaskOnly{
			TaskType:     issue.TaskType,
			CustomFields: formatIssueCustomFields(issue, pb.PropertyIssueTypeEnum_TASK, data),
		}
		model.BugOnly = IssueSheetModelBugOnly{
			OwnerName:    data.getUserNick(issue.Owner),
			Source:       issue.Source,
			ReopenCount:  issue.ReopenCount,
			CustomFields: formatIssueCustomFields(issue, pb.PropertyIssueTypeEnum_BUG, data),
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

func (data DataForFulfill) getUserNick(userid string) string {
	if userid == "" {
		return ""
	}
	if u, ok := data.UserMap[userid]; ok {
		return u.Nick
	}
	return ""
}
