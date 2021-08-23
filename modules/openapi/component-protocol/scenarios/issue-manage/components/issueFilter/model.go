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

package issueFilter

import (
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

type ComponentFilter struct {
	CtxBdl protocol.ContextBundle `json:"-"`
	filter.CommonFilter
	State    State    `json:"state,omitempty"`
	InParams InParams `json:"-"`
}

// FrontendConditions 前端支持的过滤参数
type FrontendConditions struct {
	IterationIDs       []int64                       `json:"iterationIDs,omitempty"`
	Title              string                        `json:"title,omitempty"`
	StateBelongs       []apistructs.IssueStateBelong `json:"stateBelongs,omitempty"`
	LabelIDs           []uint64                      `json:"labelIDs,omitempty"`
	Priorities         []apistructs.IssuePriority    `json:"priorities,omitempty"`
	Severities         []apistructs.IssueSeverity    `json:"severities,omitempty"`
	CreatorIDs         []string                      `json:"creatorIDs,omitempty"`
	AssigneeIDs        []string                      `json:"assigneeIDs,omitempty"`
	OwnerIDs           []string                      `json:"ownerIDs,omitempty"`
	BugStages          []string                      `json:"bugStages,omitempty"`
	CreatedAtStartEnd  []*int64                      `json:"createdAtStartEnd,omitempty"`
	FinishedAtStartEnd []*int64                      `json:"finishedAtStartEnd,omitempty"`
	ClosedAtStartEnd   []*int64                      `json:"closedAtStartEnd,omitempty"`
}

func generateFrontendConditionProps(fixedIssueType string, state State) FrontendConditionProps {
	conditionProps := []filter.PropCondition{
		{
			Key:         PropConditionKeyIterationIDs,
			Label:       "迭代",
			EmptyText:   "全部",
			Fixed:       true,
			ShowIndex:   1,
			HaveFilter:  true,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: "选择迭代",
			Options:     nil,
		},
		{
			Key:         PropConditionKeyTitle,
			Label:       "标题",
			EmptyText:   "全部",
			Fixed:       true,
			ShowIndex:   2,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeInput,
			Placeholder: "请输入标题或ID",
		},
		{
			Key:         PropConditionKeyLabelIDs,
			Label:       "标签",
			EmptyText:   "全部",
			Fixed:       false,
			ShowIndex:   0,
			HaveFilter:  true,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: "选择标签",
			Options:     nil,
		},
		{
			Key:         PropConditionKeyPriorities,
			Label:       "优先级",
			EmptyText:   "全部",
			Fixed:       false,
			ShowIndex:   0,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: "选择优先级",
			Options: []filter.PropConditionOption{
				{Label: "紧急", Value: "URGENT", Icon: ""},
				{Label: "高", Value: "HIGH", Icon: ""},
				{Label: "中", Value: "NORMAL", Icon: ""},
				{Label: "低", Value: "LOW", Icon: ""},
			},
		},
		{
			Key:         PropConditionKeySeverities,
			Label:       "严重程度",
			EmptyText:   "全部",
			Fixed:       false,
			ShowIndex:   0,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: "选择优先级",
			Options: []filter.PropConditionOption{
				{Label: "致命", Value: "FATAL", Icon: ""},
				{Label: "严重", Value: "SERIOUS", Icon: ""},
				{Label: "一般", Value: "NORMAL", Icon: ""},
				{Label: "轻微", Value: "SLIGHT", Icon: ""},
				{Label: "建议", Value: "SUGGEST", Icon: ""},
			},
		},
		{
			Key:        PropConditionKeyCreatorIDs,
			Label:      "创建人",
			EmptyText:  "全部",
			Fixed:      false,
			ShowIndex:  0,
			HaveFilter: true,
			Type:       filter.PropConditionTypeSelect,
			QuickSelect: filter.QuickSelect{
				Label:        "选择自己",
				OperationKey: OperationKeyCreatorSelectMe,
			},
			Placeholder: "",
			Options:     nil,
		},
		{
			Key:        PropConditionKeyAssigneeIDs,
			Label:      "处理人",
			EmptyText:  "全部",
			Fixed:      false,
			ShowIndex:  0,
			HaveFilter: true,
			Type:       filter.PropConditionTypeSelect,
			QuickSelect: filter.QuickSelect{
				Label:        "选择自己",
				OperationKey: OperationKeyAssigneeSelectMe,
			},
			Placeholder: "",
			Options:     nil,
		},
		{
			Key:        PropConditionKeyOwnerIDs,
			Label:      "责任人",
			EmptyText:  "全部",
			Fixed:      false,
			ShowIndex:  0,
			HaveFilter: true,
			Type:       filter.PropConditionTypeSelect,
			QuickSelect: filter.QuickSelect{
				Label:        "选择自己",
				OperationKey: OperationKeyOwnerSelectMe,
			},
			Placeholder: "",
			Options:     nil,
		},
		{
			Key: PropConditionKeyBugStages,
			Label: func() string {
				switch fixedIssueType {
				case "ALL":
				case apistructs.IssueTypeEpic.String():
				case apistructs.IssueTypeRequirement.String():
				case apistructs.IssueTypeTask.String():
					return "任务类型"
				case apistructs.IssueTypeBug.String():
					return "引入源"
				}
				return ""
			}(),
			EmptyText:   "全部",
			Fixed:       false,
			HaveFilter:  false,
			ShowIndex:   0,
			Type:        filter.PropConditionTypeSelect,
			QuickSelect: filter.QuickSelect{},
			Placeholder: "",
			Options: []filter.PropConditionOption{
				{Label: "需求设计", Value: "demandDesign", Icon: ""},
				{Label: "架构设计", Value: "architectureDesign", Icon: ""},
				{Label: "代码研发", Value: "codeDevelopment", Icon: ""},
			},
		},
		{
			Key:         PropConditionKeyCreatedAtStartEnd,
			Label:       "创建日期",
			EmptyText:   "全部",
			Fixed:       false,
			ShowIndex:   0,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeDateRange,
			QuickSelect: filter.QuickSelect{},
			Placeholder: "",
			Options:     nil,
			CustomProps: map[string]interface{}{
				"borderTime": true,
			},
		},
		{
			Key:         PropConditionKeyFinishedAtStartEnd,
			Label:       "截止日期",
			EmptyText:   "",
			Fixed:       false,
			ShowIndex:   0,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeDateRange,
			QuickSelect: filter.QuickSelect{},
			Placeholder: "",
			Options:     nil,
			CustomProps: map[string]interface{}{
				"borderTime": true,
			},
		},
	}

	if fixedIssueType == apistructs.IssueTypeBug.String() {
		conditionProps = append(conditionProps, filter.PropCondition{
			Key:         PropConditionKeyClosed,
			Label:       "关闭日期",
			EmptyText:   "",
			Fixed:       false,
			ShowIndex:   0,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeDateRange,
			QuickSelect: filter.QuickSelect{},
			Placeholder: "",
			Options:     nil,
			CustomProps: map[string]interface{}{
				"borderTime": true,
			},
		})
	}

	v, ok := state.IssueViewGroupChildrenValue["kanban"]
	if state.IssueViewGroupValue != "kanban" || !ok || v != "status" {
		status := filter.PropCondition{
			Key:         PropConditionKeyStateBelongs,
			Label:       "状态",
			EmptyText:   "全部",
			Fixed:       true,
			ShowIndex:   3,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: "",
			Options: func() []filter.PropConditionOption {
				open := filter.PropConditionOption{Label: "待处理", Value: apistructs.IssueStateBelongOpen, Icon: ""}
				reopen := filter.PropConditionOption{Label: "重新打开", Value: apistructs.IssueStateBelongReopen, Icon: ""}
				resolved := filter.PropConditionOption{Label: "已解决", Value: apistructs.IssueStateBelongResloved, Icon: ""}
				wontfix := filter.PropConditionOption{Label: "不修复", Value: apistructs.IssueStateBelongWontfix, Icon: ""}
				closed := filter.PropConditionOption{Label: "已关闭", Value: apistructs.IssueStateBelongClosed, Icon: ""}
				working := filter.PropConditionOption{Label: "进行中", Value: apistructs.IssueStateBelongWorking, Icon: ""}
				done := filter.PropConditionOption{Label: "已完成", Value: apistructs.IssueStateBelongDone, Icon: ""}
				switch fixedIssueType {
				case "ALL":
					return []filter.PropConditionOption{open, working, done, reopen, resolved, wontfix, closed}
				case apistructs.IssueTypeEpic.String():
					return []filter.PropConditionOption{open, working, done}
				case apistructs.IssueTypeRequirement.String():
					return []filter.PropConditionOption{open, working, done}
				case apistructs.IssueTypeTask.String():
					return []filter.PropConditionOption{open, working, done}
				case apistructs.IssueTypeBug.String():
					return []filter.PropConditionOption{open, working, wontfix, reopen, resolved, closed}
				}
				return nil
			}(),
		}
		conditionProps = append(conditionProps[:2], append([]filter.PropCondition{status}, conditionProps[2:]...)...)
	}

	return conditionProps
}

var (
	PropConditionKeyIterationIDs       filter.PropConditionKey = "iterationIDs"
	PropConditionKeyTitle              filter.PropConditionKey = "title"
	PropConditionKeyStateBelongs       filter.PropConditionKey = "stateBelongs"
	PropConditionKeyLabelIDs           filter.PropConditionKey = "labelIDs"
	PropConditionKeyPriorities         filter.PropConditionKey = "priorities"
	PropConditionKeySeverities         filter.PropConditionKey = "severities"
	PropConditionKeyCreatorIDs         filter.PropConditionKey = "creatorIDs"
	PropConditionKeyAssigneeIDs        filter.PropConditionKey = "assigneeIDs"
	PropConditionKeyOwnerIDs           filter.PropConditionKey = "ownerIDs"
	PropConditionKeyBugStages          filter.PropConditionKey = "bugStages"
	PropConditionKeyCreatedAtStartEnd  filter.PropConditionKey = "createdAtStartEnd"
	PropConditionKeyFinishedAtStartEnd filter.PropConditionKey = "finishedAtStartEnd"
	PropConditionKeyClosed             filter.PropConditionKey = "closedAtStartEnd"
)

func GetAllOperations() map[filter.OperationKey]filter.Operation {
	var allOperations = map[filter.OperationKey]filter.Operation{
		OperationKeyFilter:           {Key: OperationKeyFilter, Reload: true},
		OperationKeyAssigneeSelectMe: {Key: OperationKeyAssigneeSelectMe, Reload: true},
		OperationKeyCreatorSelectMe:  {Key: OperationKeyCreatorSelectMe, Reload: true},
		OperationKeyOwnerSelectMe:    {Key: OperationKeyOwnerSelectMe, Reload: true},
	}
	return allOperations
}

var (
	OperationKeyFilter           filter.OperationKey = "filter"
	OperationKeyCreatorSelectMe  filter.OperationKey = "creatorSelectMe"
	OperationKeyAssigneeSelectMe filter.OperationKey = "assigneeSelectMe"
	OperationKeyOwnerSelectMe    filter.OperationKey = "ownerSelectMe"
)
