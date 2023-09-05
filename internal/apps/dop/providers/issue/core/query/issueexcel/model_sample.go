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
	"time"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
)

// GenerateSampleIssueSheetModels
//
// one requirement, includes one task
// the bug connects to the task
func (data DataForFulfill) GenerateSampleIssueSheetModels() []IssueSheetModel {
	// time
	now := time.Now()
	nowPlusOneDay := now.AddDate(0, 0, 1)

	common := IssueSheetModelCommon{
		ID:                 0,
		IterationName:      "迭代名",
		IssueType:          0,
		IssueTitle:         "标题",
		Content:            "内容",
		State:              "",
		Priority:           pb.IssuePriorityEnum_NORMAL,
		Complexity:         pb.IssueComplexityEnum_NORMAL,
		Severity:           pb.IssueSeverityEnum_NORMAL,
		CreatorName:        data.UserID,
		AssigneeName:       data.UserID,
		CreatedAt:          &now,
		UpdatedAt:          &now,
		PlanStartedAt:      &now,
		PlanFinishedAt:     &nowPlusOneDay,
		StartAt:            &now,
		FinishAt:           nil,
		EstimateTime:       "",
		Labels:             []string{"label1", "label2"},
		ConnectionIssueIDs: nil,
	}
	// requirement
	requirementCommon := common
	requirementCommon.ID = 1
	requirementCommon.IssueType = pb.IssueTypeEnum_REQUIREMENT
	requirementCommon.Content = "这个一个需求，包含一个任务"
	requirementCommon.State = "进行中"
	requirementCommon.EstimateTime = "2d"
	requirement := IssueSheetModel{
		Common: requirementCommon,
		RequirementOnly: IssueSheetModelRequirementOnly{
			InclusionIssueIDs: []int64{2, -(3 + uuidPartsMustLength)},
			CustomFields:      formatIssueCustomFields(&pb.Issue{Id: int64(requirementCommon.ID)}, pb.PropertyIssueTypeEnum_REQUIREMENT, data),
		},
		// Assign to `TaskOnly` and `BugOnly` fields to avoid uuid order mismatch error.
		// Requirement is the first model, so just do for requirement only.
		TaskOnly: IssueSheetModelTaskOnly{
			TaskType:     "",
			CustomFields: formatIssueCustomFields(&pb.Issue{Id: int64(requirementCommon.ID)}, pb.PropertyIssueTypeEnum_TASK, data),
		},
		BugOnly: IssueSheetModelBugOnly{
			OwnerName:    "",
			Source:       "",
			ReopenCount:  0,
			CustomFields: formatIssueCustomFields(&pb.Issue{Id: int64(requirementCommon.ID)}, pb.PropertyIssueTypeEnum_BUG, data),
		},
	}
	// task
	taskCommon := common
	taskCommon.ID = 2
	taskCommon.IssueType = pb.IssueTypeEnum_TASK
	taskCommon.Content = "这个一个任务，被需求包含"
	taskCommon.State = "已完成"
	taskCommon.FinishAt = &nowPlusOneDay
	taskCommon.EstimateTime = "1d"
	taskCommon.Labels = []string{"label1"}
	task := IssueSheetModel{
		Common: taskCommon,
		TaskOnly: IssueSheetModelTaskOnly{
			TaskType:     "开发",
			CustomFields: formatIssueCustomFields(&pb.Issue{Id: int64(taskCommon.ID)}, pb.PropertyIssueTypeEnum_TASK, data),
		},
	}
	// bug
	bugCommon := common
	bugCommon.ID = 0
	bugCommon.IssueType = pb.IssueTypeEnum_BUG
	bugCommon.Content = "这个一个bug，关联到任务"
	bugCommon.PlanStartedAt = nil
	bugCommon.PlanFinishedAt = nil
	bugCommon.StartAt = nil
	bugCommon.State = "待处理"
	bugCommon.ConnectionIssueIDs = []int64{2}
	bugCommon.Labels = []string{"label2"}
	bug := IssueSheetModel{
		Common: bugCommon,
		BugOnly: IssueSheetModelBugOnly{
			OwnerName:    data.UserID,
			Source:       "代码研发",
			ReopenCount:  0,
			CustomFields: formatIssueCustomFields(&pb.Issue{Id: int64(bugCommon.ID)}, pb.PropertyIssueTypeEnum_BUG, data),
		},
	}

	return []IssueSheetModel{requirement, task, bug}
}
