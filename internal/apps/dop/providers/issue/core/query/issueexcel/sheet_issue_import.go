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
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/pkg/excel"
)

func (data DataForFulfill) DecodeIssueSheet(r io.Reader) ([]IssueSheetModel, error) {
	excelSheets, err := excel.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode excel, err: %v", err)
	}
	// convert [][][]string to map[uuid]excel.Column
	issueSheetRows := excelSheets[indexOfSheetIssue]
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
		}
	}()
	// prepare models
	var columnLen int
	for _, column := range m {
		columnLen = len(column)
		break
	}
	models := make([]IssueSheetModel, columnLen, columnLen)
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
					id, err := strconv.ParseInt(cell.Value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("invalid id: %s", cell.Value)
					}
					model.Common.ID = id
				case "IterationName":
					model.Common.IterationName = cell.Value
				case "IssueType":
					switch cell.Value {
					case pb.IssueTypeEnum_REQUIREMENT.String():
						model.Common.IssueType = pb.IssueTypeEnum_REQUIREMENT
					case pb.IssueTypeEnum_TASK.String():
						model.Common.IssueType = pb.IssueTypeEnum_TASK
					case pb.IssueTypeEnum_BUG.String():
						model.Common.IssueType = pb.IssueTypeEnum_BUG
					default:
						return nil, fmt.Errorf("invalid issue type: %s", cell.Value)
					}
				case "IssueTitle":
					model.Common.IssueTitle = cell.Value
				case "Content":
					model.Common.Content = cell.Value
				case "State":
					model.Common.State = cell.Value
				case "Priority":
					switch cell.Value {
					case pb.IssuePriorityEnum_HIGH.String():
						model.Common.Priority = pb.IssuePriorityEnum_HIGH
					case pb.IssuePriorityEnum_NORMAL.String():
						model.Common.Priority = pb.IssuePriorityEnum_NORMAL
					case pb.IssuePriorityEnum_LOW.String():
						model.Common.Priority = pb.IssuePriorityEnum_LOW
					default:
						return nil, fmt.Errorf("invalid priority: %s", cell.Value)
					}
				case "Complexity":
					switch cell.Value {
					case pb.IssueComplexityEnum_HARD.String():
						model.Common.Complexity = pb.IssueComplexityEnum_HARD
					case pb.IssueComplexityEnum_NORMAL.String():
						model.Common.Complexity = pb.IssueComplexityEnum_NORMAL
					case pb.IssueComplexityEnum_EASY.String():
						model.Common.Complexity = pb.IssueComplexityEnum_EASY
					default:
						return nil, fmt.Errorf("invalid complexity: %s", cell.Value)
					}
				case "Severity":
					switch cell.Value {
					case pb.IssueSeverityEnum_FATAL.String():
						model.Common.Severity = pb.IssueSeverityEnum_FATAL
					case pb.IssueSeverityEnum_SERIOUS.String():
						model.Common.Severity = pb.IssueSeverityEnum_SERIOUS
					case pb.IssueSeverityEnum_NORMAL.String():
						model.Common.Severity = pb.IssueSeverityEnum_NORMAL
					default:
						return nil, fmt.Errorf("invalid severity: %s", cell.Value)
					}
				case "CreatorName":
					model.Common.CreatorName = cell.Value
				case "AssigneeName":
					model.Common.AssigneeName = cell.Value
				case "CreatedAt":
					model.Common.CreatedAt = mustParseStringTime(cell.Value, groupField)
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
					model.Common.Labels = strings.Split(cell.Value, ",")
				case "ConnectionIssueIDs":
					var ids []int64
					for _, idStr := range strings.Split(cell.Value, ",") {
						id, err := strconv.ParseInt(idStr, 10, 64)
						if err != nil {
							return nil, fmt.Errorf("invalid connection issue id: %s", idStr)
						}
						ids = append(ids, id)
					}
					model.Common.ConnectionIssueIDs = ids
				default:
					return nil, fmt.Errorf("unknown common field: %s", groupField)
				}
			case "RequirementOnly":
				switch groupField {
				case "InclusionIssueIDs":
					var ids []int64
					for _, idStr := range strings.Split(cell.Value, ",") {
						id, err := strconv.ParseInt(idStr, 10, 64)
						if err != nil {
							return nil, fmt.Errorf("invalid inclusion issue id: %s", idStr)
						}
						ids = append(ids, id)
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
