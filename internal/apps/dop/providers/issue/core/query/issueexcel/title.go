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

//type (
//	EnhancedCell struct {
//		excel.Cell
//		UUID       string // TODO
//		DataSetter DataSetter
//	}
//	DataSetter func(cellTitle string, issue *pb.Issue, data DataForFulfill) any
//)
//
//func ConvertRawCell(rawCell excel.Cell) EnhancedCell {
//	return EnhancedCell{Cell: rawCell}
//}
//
//func BatchConvertRawCells(rawCells []excel.Cell) []EnhancedCell {
//	var results []EnhancedCell
//	for _, cell := range rawCells {
//		results = append(results, ConvertRawCell(cell))
//	}
//	return results
//}
//
//func ExtractToRawCells(setterCells []EnhancedCell) []excel.Cell {
//	var results []excel.Cell
//	for _, cell := range setterCells {
//		results = append(results, cell.Cell)
//	}
//	return results
//}

//// generateIssueSheetTitle 通过 row struct 反射生成 excel title
//// 生成的 Cell 里包含数据，每一列有且只有一个 cell 有数据，这个 cell 的 uuid 是从上往下经过的 cell 的 value 拼接 ('-')
//// 例如 任务类型，这个 cell 会有 data，且该 cell 的 uuid = IssueSheetModelTaskOnly-TaskType
//// 解析 excel 时使用同样的规则，每一个数据行的 cell，对应的 uuid，通过 - 进行 split，最终映射为 IssueSheetModel
//func generateIssueSheetTitle(data DataForFulfill) IssueSheetTitleCells {
//	//localeRes := data.Locale
//	row := IssueSheetModel{}
//	for issueType, customFields := range data.CustomFieldMap {
//		switch issueType {
//		case pb.PropertyIssueTypeEnum_REQUIREMENT:
//			for _, cf := range customFields {
//				row.RequirementOnly.CustomFields = append(row.RequirementOnly.CustomFields, ExcelCustomField{Title: cf.PropertyName})
//			}
//		case pb.PropertyIssueTypeEnum_TASK:
//			for _, cf := range customFields {
//				row.TaskOnly.CustomFields = append(row.TaskOnly.CustomFields, ExcelCustomField{Title: cf.PropertyName})
//			}
//		case pb.PropertyIssueTypeEnum_BUG:
//			for _, cf := range customFields {
//				row.BugOnly.CustomFields = append(row.BugOnly.CustomFields, ExcelCustomField{Title: cf.PropertyName})
//			}
//		default:
//			continue
//		}
//	}
//	rowValue := reflect.ValueOf(&row).Elem()
//	var excelRowCells IssueSheetTitleCells
//	for i := 0; i < rowValue.NumField(); i++ {
//		valueField := rowValue.Field(i)
//		_ = valueField
//		typeField := rowValue.Type().Field(i)
//		switch typeField.Type {
//		case reflect.TypeOf(IssueSheetModelCommon{}):
//			excelRowCells.CommonTitleCells = handleCommonTitle(data)
//		case reflect.TypeOf(IssueSheetModelRequirementOnly{}):
//			excelRowCells.RequirementOnlyCells = handleIssueTypeOnlyTitle(data, valueField)
//		case reflect.TypeOf(IssueSheetModelTaskOnly{}):
//			excelRowCells.TaskOnlyCells = handleIssueTypeOnlyTitle(data, valueField)
//		case reflect.TypeOf(IssueSheetModelBugOnly{}):
//			excelRowCells.BugOnlyCells = handleIssueTypeOnlyTitle(data, valueField)
//		default:
//			panic(fmt.Sprintf("unknown struct field: %s", typeField.Name))
//		}
//	}
//	return excelRowCells
//	//// convert to [][]excel.Cell
//	//return mergeMultiLayerTitleCells(excelRowCells)
//}
//
//// IssueSheetTitleCells 每个字段独立完成 3 行的 cells，需要填充 empty cell
//type IssueSheetTitleCells struct {
//	CommonTitleCells     [][]EnhancedCell
//	RequirementOnlyCells [][]EnhancedCell
//	TaskOnlyCells        [][]EnhancedCell
//	BugOnlyCells         [][]EnhancedCell
//}
//
//func mergeLineCells(slices ...[]EnhancedCell) []excel.Cell {
//	var result []excel.Cell
//	for _, slice := range slices {
//		result = append(result, ExtractToRawCells(slice)...)
//	}
//	return result
//}
//
//func mergeMultiLayerTitleCells(cells IssueSheetTitleCells) [][]excel.Cell {
//	// order by: common, requirement, task, bug
//	// title's layer deep is 3
//	firstLine := mergeLineCells(cells.CommonTitleCells[0], cells.RequirementOnlyCells[0], cells.TaskOnlyCells[0], cells.BugOnlyCells[0])
//	secondLine := mergeLineCells(cells.CommonTitleCells[1], cells.RequirementOnlyCells[1], cells.TaskOnlyCells[1], cells.BugOnlyCells[1])
//	thirdLine := mergeLineCells(cells.CommonTitleCells[2], cells.RequirementOnlyCells[2], cells.TaskOnlyCells[2], cells.BugOnlyCells[2])
//	return append([][]excel.Cell{}, firstLine, secondLine, thirdLine)
//}
//
//func handleCommonTitle(data DataForFulfill) [][]EnhancedCell {
//	commonType := reflect.TypeOf(IssueSheetModelCommon{})
//	var cells []EnhancedCell
//	for i := 0; i < commonType.NumField(); i++ {
//		structField := commonType.Field(i)
//		title := getFieldRawCellTitle(structField)
//		rawCell := excel.NewVMergeCell(title, 2)
//		cell := EnhancedCell{
//			Cell: rawCell,
//			DataSetter: func(cellTitle string, issue *pb.Issue, data DataForFulfill) any {
//				switch cellTitle {
//				case "ID":
//					return issue.Id
//				case "Iteration":
//					return data.IterationMapByID[issue.IterationID]
//				case "IssueType":
//					return issue.Type
//				case "IssueTitle":
//					return issue.Title
//				case "Content":
//					return issue.Content
//				case "State":
//					return data.StateMap[issue.State]
//				case "Priority":
//					return issue.Priority
//				case "Complexity":
//					return issue.Complexity
//				case "Severity":
//					return issue.Severity
//				case "CreatorName":
//					return data.ProjectMemberMap[issue.Creator]
//				case "AssigneeName":
//					return data.ProjectMemberMap[issue.Assignee]
//				case "CreatedAt":
//					return formatTimeFromTimestamp(issue.CreatedAt)
//				case "PlanStartedAt":
//					return formatTimeFromTimestamp(issue.PlanStartedAt)
//				case "PlanFinishedAt":
//					return formatTimeFromTimestamp(issue.PlanFinishedAt)
//				case "StartAt":
//					return formatTimeFromTimestamp(issue.StartTime)
//				case "FinishAt":
//					return formatTimeFromTimestamp(issue.FinishTime)
//				case "EstimateTime":
//					return streamcommon.GetFormartTime(issue.IssueManHour, "EstimateTime")
//				case "Labels":
//					return issue.Labels
//				case "ConnectionIssueIDs":
//					return data.ConnectionMap[issue.Id]
//				default:
//					panic(fmt.Sprintf("unknown common title: %s", title))
//				}
//			},
//		}
//		cells = append(cells, cell)
//	}
//	return [][]EnhancedCell{cells, BatchConvertRawCells(excel.EmptyCells(len(cells))), BatchConvertRawCells(excel.EmptyCells(len(cells)))}
//}

//// 需要传入具体的 IssueSheetModelRequirementOnly 字段实例
//func handleIssueTypeOnlyTitle(data DataForFulfill, v reflect.Value) [][]EnhancedCell {
//	// maxLen = IssueSheetModelRequirementOnly 的非 CustomFields 字段数量 + CustomFields 的数量
//	var maxLen int
//	var customFields []ExcelCustomField
//	var exclusiveFields []string
//	for i := 0; i < v.NumField(); i++ {
//		valueField := v.Field(i)
//		typeField := v.Type().Field(i)
//		switch typeField.Type {
//		case reflect.TypeOf([]ExcelCustomField{}):
//			customFields = handleCustomFieldsTitle(valueField)
//			maxLen += len(customFields)
//		default:
//			maxLen++
//			exclusiveField := getFieldI18nCellTitle(data.Locale, typeField)
//			exclusiveFields = append(exclusiveFields, exclusiveField)
//		}
//	}
//	// construct 3-level title
//	// | 需求专属字段                |
//	// | 待办事项 | 自定义字段       |
//	// |         | 字段1,字段2(动态)|
//	// first line
//	firstLine := BatchConvertRawCells(excel.NewHMergeCellsAuto(v.Type().Name(), maxLen-1))
//	// second line
//	var secondLine []EnhancedCell
//	for _, field := range exclusiveFields {
//		setterCell := ConvertRawCell(excel.NewVMergeCell(field, 1))
//		setterCell.DataSetter = func(cellTitle string, issue *pb.Issue, data DataForFulfill) any {
//			switch cellTitle {
//			case "InclusionIssueIDs":
//				return data.InclusionMap[issue.Id]
//			case "TaskType":
//				return data.StageMap[query.IssueStage{Type: issue.Type.String(), Value: common.GetStage(issue)}]
//			case "OwnerName":
//				return data.ProjectMemberMap[issue.Owner]
//			case "Source":
//				return issue.Source
//			case "ReopenCount":
//				return issue.ReopenCount
//			default:
//				panic(fmt.Errorf("unknown exclusive title: %s", cellTitle))
//			}
//		}
//		secondLine = append(secondLine, setterCell)
//	}
//	secondLine = append(secondLine, BatchConvertRawCells(excel.NewHMergeCellsAuto("CustomFields", len(customFields)-1))...)
//	// third line
//	var thirdLine []EnhancedCell
//	thirdLine = append(thirdLine, BatchConvertRawCells(excel.EmptyCells(len(exclusiveFields)))...)
//	// add customFields at last
//	var customFieldsCells []EnhancedCell
//	for _, customField := range customFields {
//		setterCell := ConvertRawCell(excel.NewCell(customField.Title))
//		setterCell.DataSetter = func(cellTitle string, issue *pb.Issue, data DataForFulfill) any {
//			return formatOneCustomField(customField.Origin, issue, data)
//		}
//		customFieldsCells = append(customFieldsCells, setterCell)
//	}
//	thirdLine = append(thirdLine, customFieldsCells...)
//	return [][]EnhancedCell{firstLine, secondLine, thirdLine}
//}

//func handleCustomFieldsTitle(valueField reflect.Value) []ExcelCustomField {
//	var customFields []ExcelCustomField
//	for i := 0; i < valueField.Len(); i++ {
//		customFieldValue := valueField.Index(i)
//		customField := customFieldValue.Interface().(ExcelCustomField)
//		customFields = append(customFields, customField)
//	}
//	return customFields
//}

//func getFieldRawCellTitle(field reflect.StructField) string {
//	fieldName := field.Name
//	tag := field.Tag.Get("excel")
//	if tag == "" {
//		tag = fieldName
//	}
//	return tag
//}
//
//func getFieldI18nCellTitle(localeRes *i18n.LocaleResource, field reflect.StructField) string {
//	rawTitle := getFieldRawCellTitle(field)
//	i18nCellTitle := localeRes.Get(rawTitle, rawTitle)
//	return i18nCellTitle
//}
