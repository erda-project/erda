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
	"testing"
)

func Test_generateExcelTitleByRowStruct(t *testing.T) {
	//data := DataForFulfill{
	//	Locale: &i18n.LocaleResource{},
	//	CustomFieldMap: map[pb.PropertyIssueTypeEnum_PropertyIssueType][]*pb.IssuePropertyIndex{
	//		pb.PropertyIssueTypeEnum_REQUIREMENT: []*pb.IssuePropertyIndex{
	//			{
	//				PropertyName:      "需求自定义字段-1",
	//				PropertyIssueType: pb.PropertyIssueTypeEnum_REQUIREMENT,
	//				Index:             0,
	//			},
	//		},
	//		pb.PropertyIssueTypeEnum_TASK: []*pb.IssuePropertyIndex{
	//			{
	//				PropertyName:      "任务自定义字段-1",
	//				PropertyIssueType: pb.PropertyIssueTypeEnum_TASK,
	//				Index:             1,
	//			},
	//			{
	//				PropertyName:      "任务自定义字段-2",
	//				PropertyIssueType: pb.PropertyIssueTypeEnum_TASK,
	//				Index:             2,
	//			},
	//		},
	//		pb.PropertyIssueTypeEnum_BUG: []*pb.IssuePropertyIndex{
	//			{
	//				PropertyName:      "缺陷自定义字段-1",
	//				PropertyIssueType: pb.PropertyIssueTypeEnum_BUG,
	//				Index:             3,
	//			},
	//			{
	//				PropertyName:      "缺陷自定义字段-2",
	//				PropertyIssueType: pb.PropertyIssueTypeEnum_BUG,
	//				Index:             4,
	//			},
	//			{
	//				PropertyName:      "缺陷自定义字段-3",
	//				PropertyIssueType: pb.PropertyIssueTypeEnum_BUG,
	//				Index:             5,
	//			},
	//		},
	//	},
	//	Issues: []*pb.Issue{
	//		{
	//			Id:                123,
	//			CreatedAt:         nil,
	//			UpdatedAt:         nil,
	//			PlanStartedAt:     nil,
	//			PlanFinishedAt:    nil,
	//			ProjectID:         0,
	//			IterationID:       0,
	//			AppID:             0,
	//			RequirementID:     0,
	//			RequirementTitle:  "",
	//			Title:             "",
	//			Content:           "",
	//			State:             0,
	//			Priority:          0,
	//			Complexity:        0,
	//			Severity:          0,
	//			Assignee:          "",
	//			Creator:           "",
	//			IssueButton:       nil,
	//			IssueSummary:      nil,
	//			Labels:            nil,
	//			LabelDetails:      nil,
	//			IssueManHour:      nil,
	//			Source:            "",
	//			TaskType:          "",
	//			BugStage:          "",
	//			Owner:             "",
	//			Subscribers:       nil,
	//			FinishTime:        nil,
	//			TestPlanCaseRels:  nil,
	//			RelatedIssueIDs:   nil,
	//			ReopenCount:       0,
	//			Type:              0,
	//			PropertyInstances: nil,
	//			StartTime:         nil,
	//		},
	//	},
	//}
	//// gen title
	//titleCellsWithDataSetter := generateIssueSheetTitle(data)
	//sheetTitleLines := mergeMultiLayerTitleCells(titleCellsWithDataSetter)
	//// gen data
	//var dataLines [][]excel.Cell
	//for _, issue := range data.Issues {
	//	row := fulfillIssueSheetDataRow(issue, data, titleCellsWithDataSetter)
	//	dataLines = append(dataLines, row)
	//}
	//// merge lines
	//allRows := append(sheetTitleLines, dataLines...)
	//
	//// gen excel file
	//f, err := os.OpenFile("./gen.xlsx", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	//assert.NoError(t, err)
	//defer f.Close()
	//err = excel.ExportExcelByCell(f, allRows, "issues")
	//assert.NoError(t, err)
}
