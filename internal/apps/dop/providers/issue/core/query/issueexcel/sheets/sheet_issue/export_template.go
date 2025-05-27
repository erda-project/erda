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
	"time"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
)

// issue-excel-template
const (
	// Template
	TemplateRequirementContent = "Template.RequirementContent"
	TemplateTaskContent        = "Template.TaskContent"
	TemplateBugContent         = "Template.BugContent"
	TemplateTicketContent      = "Template.TicketContent"
)

// GenerateSampleIssueSheetModels
//
// one requirement, includes one task
// the bug connects to the task
func GenerateSampleIssueSheetModels(data *vars.DataForFulfill) []vars.IssueSheetModel {
	// time
	now := time.Now()
	nowPlusOneDay := now.AddDate(0, 0, 1)

	userName := data.GetOrgUserNameByID(data.UserID)

	common := vars.IssueSheetModelCommon{
		ID:                 0,
		IterationName:      data.IterationMapByID[-1].Title,
		IssueType:          0,
		IssueTitle:         data.I18n(fieldIssueTitle),
		Content:            data.I18n(fieldContent),
		State:              "",
		Priority:           pb.IssuePriorityEnum_NORMAL,
		Complexity:         pb.IssueComplexityEnum_NORMAL,
		Severity:           pb.IssueSeverityEnum_NORMAL,
		CreatorName:        userName,
		AssigneeName:       userName,
		CreatedAt:          &now,
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
	requirementCommon.Content = data.I18n(TemplateRequirementContent)
	requirementCommon.State = "进行中"
	requirementCommon.EstimateTime = "2d"
	requirement := vars.IssueSheetModel{
		Common: requirementCommon,
		RequirementOnly: vars.IssueSheetModelRequirementOnly{
			InclusionIssueIDs: []int64{2, -(3 + uuidPartsMustLength)},
			CustomFields:      vars.FormatIssueCustomFields(&pb.Issue{Id: int64(requirementCommon.ID)}, pb.PropertyIssueTypeEnum_REQUIREMENT, data),
		},
	}
	// task
	taskCommon := common
	taskCommon.ID = 2
	taskCommon.IssueType = pb.IssueTypeEnum_TASK
	taskCommon.Content = data.I18n(TemplateTaskContent)
	taskCommon.State = "已完成"
	taskCommon.FinishAt = &nowPlusOneDay
	taskCommon.EstimateTime = "1d"
	taskCommon.Labels = []string{"label1"}
	var taskType string
	for kv, name := range data.StageMap {
		if kv.Type == taskCommon.IssueType.String() {
			taskType = name
			break
		}
	}
	task := vars.IssueSheetModel{
		Common: taskCommon,
		TaskOnly: vars.IssueSheetModelTaskOnly{
			TaskType:     taskType,
			CustomFields: vars.FormatIssueCustomFields(&pb.Issue{Id: int64(taskCommon.ID)}, pb.PropertyIssueTypeEnum_TASK, data),
		},
	}
	// bug
	bugCommon := common
	bugCommon.ID = 0
	bugCommon.IssueType = pb.IssueTypeEnum_BUG
	bugCommon.Content = data.I18n(TemplateBugContent)
	bugCommon.PlanStartedAt = nil
	bugCommon.PlanFinishedAt = nil
	bugCommon.StartAt = nil
	bugCommon.State = "待处理"
	bugCommon.ConnectionIssueIDs = []int64{2}
	bugCommon.Labels = []string{"label2"}
	var bugSource string
	for kv, name := range data.StageMap {
		if kv.Type == bugCommon.IssueType.String() {
			bugSource = name
			break
		}
	}
	bug := vars.IssueSheetModel{
		Common: bugCommon,
		BugOnly: vars.IssueSheetModelBugOnly{
			OwnerName:    userName,
			Source:       bugSource,
			ReopenCount:  0,
			CustomFields: vars.FormatIssueCustomFields(&pb.Issue{Id: int64(bugCommon.ID)}, pb.PropertyIssueTypeEnum_BUG, data),
		},
	}
	// ticket
	ticketCommon := common
	ticketCommon.ID = 0
	ticketCommon.IssueType = pb.IssueTypeEnum_TICKET
	ticketCommon.Content = data.I18n(TemplateTicketContent)
	ticketCommon.PlanStartedAt = nil
	ticketCommon.PlanFinishedAt = nil
	ticketCommon.StartAt = nil
	ticketCommon.State = "待处理"
	ticketCommon.ConnectionIssueIDs = []int64{2}
	ticketCommon.Labels = []string{"label2"}
	ticket := vars.IssueSheetModel{
		Common: ticketCommon,
	}

	return []vars.IssueSheetModel{requirement, task, bug, ticket}
}
