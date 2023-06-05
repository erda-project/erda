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

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/pkg/excel"
)

func (data DataForFulfill) convertOldIssueSheet(sheet [][]string) ([]IssueSheetModel, error) {
	// convert by column fixed index
	m := make(map[IssueSheetColumnUUID]excel.Column)
	addM := func(m map[IssueSheetColumnUUID]excel.Column, uuid IssueSheetColumnUUID, s string) {
		uuid.AutoComplete()
		m[uuid] = append(m[uuid], excel.NewCell(s))
	}
	for rowIdx, row := range sheet {
		if rowIdx == 0 {
			continue
		}
		for columnIdx, s := range row {
			var model IssueSheetModel
			// get issue type first
			issueType := sheet[rowIdx][13]
			switch issueType {
			case pb.IssueTypeEnum_REQUIREMENT.String(), data.Locale.Get(pb.IssueTypeEnum_REQUIREMENT.String()):
				model.Common.IssueType = pb.IssueTypeEnum_REQUIREMENT
			case pb.IssueTypeEnum_TASK.String(), data.Locale.Get(pb.IssueTypeEnum_TASK.String()):
				model.Common.IssueType = pb.IssueTypeEnum_TASK
			case pb.IssueTypeEnum_BUG.String(), data.Locale.Get(pb.IssueTypeEnum_BUG.String()):
				model.Common.IssueType = pb.IssueTypeEnum_BUG
			default:
				return nil, fmt.Errorf("invalid issue type: %s", issueType)
			}
			switch columnIdx {
			case 0: // ID
				addM(m, NewIssueSheetColumnUUID("Common", "ID"), s)
			case 1: // Title
				addM(m, NewIssueSheetColumnUUID("Common", "Title"), s)
			case 2: // Content
				addM(m, NewIssueSheetColumnUUID("Common", "Content"), s)
			case 3: // State
				addM(m, NewIssueSheetColumnUUID("Common", "State"), s)
			case 4: // Creator
				addM(m, NewIssueSheetColumnUUID("Common", "Creator"), s)
			case 5: // Assignee
				addM(m, NewIssueSheetColumnUUID("Common", "Assignee"), s)
			case 6: // Owner
				addM(m, NewIssueSheetColumnUUID("BugOnly", "Owner"), s)
			case 7: // TaskType or BugSource
				switch model.Common.IssueType {
				case pb.IssueTypeEnum_TASK:
					addM(m, NewIssueSheetColumnUUID("TaskOnly", "TaskType"), s)
				case pb.IssueTypeEnum_BUG:
					addM(m, NewIssueSheetColumnUUID("BugOnly", "Source"), s)
				}
			case 8: // Priority
				addM(m, NewIssueSheetColumnUUID("Common", "Priority"), s)
			case 9: // IterationName
				addM(m, NewIssueSheetColumnUUID("Common", "IterationName"), s)
			case 10: // Complexity
				addM(m, NewIssueSheetColumnUUID("Common", "Complexity"), s)
			case 11: // Severity
				addM(m, NewIssueSheetColumnUUID("Common", "Severity"), s)
			case 12: // Labels
				addM(m, NewIssueSheetColumnUUID("Common", "Labels"), s)
			case 13: // IssueType
				addM(m, NewIssueSheetColumnUUID("Common", "IssueType"), s)
			case 14: // PlanFinishedAt
				addM(m, NewIssueSheetColumnUUID("Common", "PlanFinishedAt"), s)
			case 15: // CreatedAt
				addM(m, NewIssueSheetColumnUUID("Common", "CreatedAt"), s)
			case 16: // ConnectionIssueIDs
				addM(m, NewIssueSheetColumnUUID("Common", "ConnectionIssueIDs"), s)
			case 17: // EstimateTime
				addM(m, NewIssueSheetColumnUUID("Common", "EstimateTime"), s)
			case 18: // FinishedAt
				addM(m, NewIssueSheetColumnUUID("Common", "FinishAt"), s)
			case 19: // StartAt
				addM(m, NewIssueSheetColumnUUID("Common", "PlanStartedAt"), s)
			case 20: // ReopenCount
				addM(m, NewIssueSheetColumnUUID("Common", "ReopenCount"), s)
			default:
				// handle custom fields
				cfName := sheet[0][columnIdx]
			CF_LOOP:
				for cfType, cfs := range data.CustomFieldMap {
					switch cfType {
					case pb.PropertyIssueTypeEnum_REQUIREMENT:
						for _, cf := range cfs {
							if cf.PropertyName == cfName {
								addM(m, NewIssueSheetColumnUUID("RequirementOnly", "CustomFields", cfName), s)
								break CF_LOOP
							}
						}
					case pb.PropertyIssueTypeEnum_TASK:
						for _, cf := range cfs {
							if cf.PropertyName == cfName {
								addM(m, NewIssueSheetColumnUUID("TaskOnly", "CustomFields", cfName), s)
								break CF_LOOP
							}
						}
					case pb.PropertyIssueTypeEnum_BUG:
						for _, cf := range cfs {
							if cf.PropertyName == cfName {
								addM(m, NewIssueSheetColumnUUID("BugOnly", "CustomFields", cfName), s)
								break CF_LOOP
							}
						}
					}
				}
			}
		}
	}
	models, err := data.decodeMapToIssueSheetModel(m)
	if err != nil {
		return nil, fmt.Errorf("failed to decode old excel format map to issue sheet model, err: %v", err)
	}
	return models, nil
}
