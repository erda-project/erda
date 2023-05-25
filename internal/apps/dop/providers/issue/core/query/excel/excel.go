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

package excel

import (
	"fmt"
	"reflect"
	"time"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/i18n"
)

// excel format:
//
// issues sheet:
// 通用字段 | 需求专属字段               | 任务专属字段             | 缺陷专属字段
// -       | 待办事项 | 自定义字段       | 任务类型  | 自定义字段   | 负责人 | 引入源 | 重新打开次数 | 自定义字段
// -                 | 字段1,字段2(动态) |          | 字段1,字段2  |       |        |             | 字段1,字段2
//
// 自定义字段 sheet
// 事项类型 | JSON
//
// 标签 sheet
// name | color
//
// 用户信息 sheet
// ID | Nick | UserInfo(JSON)
type (
	IssueExcelSheetRow struct {
		Common          ExcelCommonTitle `excel:",recursive"`
		RequirementOnly RequirementOnly  `excel:",recursive"`
		TaskOnly        TaskOnly         `excel:",recursive"`
		BugOnly         BugOnly          `excel:",recursive"`
	}
	ExcelCommonTitle struct {
		ID                 string `excel:"ID"`
		IterationName      string `excel:"Iteration"`
		IssueType          string
		IssueTitle         string
		Content            string
		State              string
		Priority           string
		Complexity         string
		Severity           string
		CreatorName        string
		AssigneeName       string
		CreatedAt          time.Time
		PlanStartedAt      time.Time
		PlanFinishedAt     time.Time
		StartAt            time.Time
		FinishAt           time.Time
		EstimateTime       time.Duration
		Labels             []struct{ Name, Color string }
		ConnectionIssueIDs []uint64
	}
	RequirementOnly struct {
		InclusionIssueIDs []uint64           `excel:"Inclusion"`
		CustomFields      []ExcelCustomField `excel:",expandCustomFields"`
	}
	TaskOnly struct {
		TaskType     string
		CustomFields []ExcelCustomField `excel:",expandCustomFields"`
	}
	BugOnly struct {
		OwnerName    string
		Source       string
		ReopenCount  int
		CustomFields []ExcelCustomField `excel:",expandCustomFields"`
	}
	ExcelCustomField struct {
		Title string
		Value string
	}
)

// generateExcelTitleByRowStruct 通过 row struct 反射生成 excel title
func generateExcelTitleByRowStruct(localeRes i18n.LocaleResource, customFields []*pb.IssuePropertyIndex) [][]excel.Cell {
	// 目标是 IssueExcelSheetTitle
	row := IssueExcelSheetRow{}
	// convert customFields
	for _, cf := range customFields {
		switch cf.PropertyIssueType {
		case pb.PropertyIssueTypeEnum_REQUIREMENT:
			row.RequirementOnly.CustomFields = append(row.RequirementOnly.CustomFields, ExcelCustomField{Title: cf.PropertyName})
		case pb.PropertyIssueTypeEnum_TASK:
			row.TaskOnly.CustomFields = append(row.TaskOnly.CustomFields, ExcelCustomField{Title: cf.PropertyName})
		case pb.PropertyIssueTypeEnum_BUG:
			row.BugOnly.CustomFields = append(row.BugOnly.CustomFields, ExcelCustomField{Title: cf.PropertyName})
		default:
			continue
		}
	}
	rowValue := reflect.ValueOf(&row).Elem()
	var excelRowCells ExcelRowCells
	for i := 0; i < rowValue.NumField(); i++ {
		valueField := rowValue.Field(i)
		typeField := rowValue.Type().Field(i)
		switch typeField.Type {
		case reflect.TypeOf(ExcelCommonTitle{}):
			excelRowCells.CommonTitleCells = handleCommonTitle(localeRes)
		case reflect.TypeOf(RequirementOnly{}):
			excelRowCells.RequirementOnlyCells = handleIssueTypeOnlyTitle(localeRes, valueField)
		case reflect.TypeOf(TaskOnly{}):
			excelRowCells.TaskOnlyCells = handleIssueTypeOnlyTitle(localeRes, valueField)
		case reflect.TypeOf(BugOnly{}):
			excelRowCells.BugOnlyCells = handleIssueTypeOnlyTitle(localeRes, valueField)
		default:
			panic(fmt.Sprintf("unknown struct field: %s", typeField.Name))
		}
	}
	// convert to [][]excel.Cell
	return mergeMultiLayerCells(excelRowCells)
}

// ExcelRowCells 每个字段独立完成 3 行的 cells，不需要填充 empty cell，merge 时会自动补齐
type ExcelRowCells struct {
	CommonTitleCells     [][]excel.Cell
	RequirementOnlyCells [][]excel.Cell
	TaskOnlyCells        [][]excel.Cell
	BugOnlyCells         [][]excel.Cell
}

func mergeLineCells(slices ...[]excel.Cell) []excel.Cell {
	var result []excel.Cell
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}

func mergeMultiLayerCells(cells ExcelRowCells) [][]excel.Cell {
	// order by: common, requirement, task, bug
	// title's layer deep is 3
	firstLine := mergeLineCells(cells.CommonTitleCells[0], cells.RequirementOnlyCells[0], cells.TaskOnlyCells[0], cells.BugOnlyCells[0])
	secondLine := mergeLineCells(cells.CommonTitleCells[1], cells.RequirementOnlyCells[1], cells.TaskOnlyCells[1], cells.BugOnlyCells[1])
	thirdLine := mergeLineCells(cells.CommonTitleCells[2], cells.RequirementOnlyCells[2], cells.TaskOnlyCells[2], cells.BugOnlyCells[2])
	return append([][]excel.Cell{}, firstLine, secondLine, thirdLine)
}

func handleCommonTitle(localeRes i18n.LocaleResource) [][]excel.Cell {
	commonType := reflect.TypeOf(ExcelCommonTitle{})
	var cells []excel.Cell
	for i := 0; i < commonType.NumField(); i++ {
		field := commonType.Field(i)
		title := getFieldI18nCellTitle(localeRes, field)
		cells = append(cells, excel.NewVMergeCell(title, 2))
	}
	return [][]excel.Cell{cells, excel.EmptyCells(len(cells)), excel.EmptyCells(len(cells))}
}

// 需要传入具体的 RequirementOnly 字段实例
func handleIssueTypeOnlyTitle(localeRes i18n.LocaleResource, v reflect.Value) [][]excel.Cell {
	// maxLen = RequirementOnly 的非 CustomFields 字段数量 + CustomFields 的数量
	var maxLen int
	var customFields []ExcelCustomField
	var exclusiveFields []string
	for i := 0; i < v.NumField(); i++ {
		valueField := v.Field(i)
		typeField := v.Type().Field(i)
		switch typeField.Type {
		case reflect.TypeOf([]ExcelCustomField{}):
			customFields = handleCustomFieldsTitle(valueField)
			maxLen += len(customFields)
		default:
			maxLen++
			exclusiveField := getFieldI18nCellTitle(localeRes, typeField)
			exclusiveFields = append(exclusiveFields, exclusiveField)
		}
	}
	// construct 3-level title
	// | 需求专属字段                |
	// | 待办事项 | 自定义字段       |
	// |         | 字段1,字段2(动态)|
	// first line
	firstLine := excel.NewHMergeCellsAuto(v.Type().Name(), maxLen-1)
	// second line
	var secondLine []excel.Cell
	for _, field := range exclusiveFields {
		secondLine = append(secondLine, excel.NewVMergeCell(field, 1))
	}
	secondLine = append(secondLine, excel.NewHMergeCellsAuto("CustomFields", len(customFields)-1)...)
	// third line
	var thirdLine []excel.Cell
	thirdLine = append(thirdLine, excel.EmptyCells(len(exclusiveFields))...)
	// add customFields at last
	var customFieldsCells []excel.Cell
	for _, customField := range customFields {
		customFieldsCells = append(customFieldsCells, excel.NewCell(customField.Title))
	}
	thirdLine = append(thirdLine, customFieldsCells...)
	return [][]excel.Cell{firstLine, secondLine, thirdLine}
}

func handleCustomFieldsTitle(valueField reflect.Value) []ExcelCustomField {
	var customFields []ExcelCustomField
	for i := 0; i < valueField.Len(); i++ {
		customFieldValue := valueField.Index(i)
		customField := customFieldValue.Interface().(ExcelCustomField)
		customFields = append(customFields, customField)
	}
	return customFields
}

func getFieldI18nCellTitle(localeRes i18n.LocaleResource, field reflect.StructField) string {
	fieldName := field.Name
	tag := field.Tag.Get("excel")
	if tag == "" {
		tag = fieldName
	}
	i18nCellTitle := localeRes.Get(tag, tag)
	return i18nCellTitle
}
