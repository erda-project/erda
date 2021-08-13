// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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

func generateFrontendConditionProps(fixedIssueType string, state State, bdl protocol.ContextBundle) FrontendConditionProps {
	conditionProps := []filter.PropCondition{
		{
			Key:         PropConditionKeyIterationIDs,
			Label:       bdl.I18nPrinter.Sprintf("Sprint"),
			EmptyText:   bdl.I18nPrinter.Sprintf("All"),
			Fixed:       true,
			ShowIndex:   1,
			HaveFilter:  true,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: bdl.I18nPrinter.Sprintf("Choose Sprint"),
			Options:     nil,
		},
		{
			Key:         PropConditionKeyTitle,
			Label:       bdl.I18nPrinter.Sprintf("Title"),
			EmptyText:   bdl.I18nPrinter.Sprintf("All"),
			Fixed:       true,
			ShowIndex:   2,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeInput,
			Placeholder: bdl.I18nPrinter.Sprintf("Please enter title or ID"),
		},
		{
			Key:         PropConditionKeyLabelIDs,
			Label:       bdl.I18nPrinter.Sprintf("Label"),
			EmptyText:   bdl.I18nPrinter.Sprintf("All"),
			Fixed:       false,
			ShowIndex:   0,
			HaveFilter:  true,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: bdl.I18nPrinter.Sprintf("Please choose label"),
			Options:     nil,
		},
		{
			Key:         PropConditionKeyPriorities,
			Label:       bdl.I18nPrinter.Sprintf("Priority"),
			EmptyText:   bdl.I18nPrinter.Sprintf("All"),
			Fixed:       false,
			ShowIndex:   0,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: bdl.I18nPrinter.Sprintf("Choose Priorities"),
			Options: []filter.PropConditionOption{
				{Label: bdl.I18nPrinter.Sprintf("URGENT"), Value: "URGENT", Icon: ""},
				{Label: bdl.I18nPrinter.Sprintf("HIGH"), Value: "HIGH", Icon: ""},
				{Label: bdl.I18nPrinter.Sprintf("NORMAL"), Value: "NORMAL", Icon: ""},
				{Label: bdl.I18nPrinter.Sprintf("LOW"), Value: "LOW", Icon: ""},
			},
		},
		{
			Key:         PropConditionKeySeverities,
			Label:       bdl.I18nPrinter.Sprintf("Severity"),
			EmptyText:   bdl.I18nPrinter.Sprintf("All"),
			Fixed:       false,
			ShowIndex:   0,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: bdl.I18nPrinter.Sprintf("Choose Severity"),
			Options: []filter.PropConditionOption{
				{Label: bdl.I18nPrinter.Sprintf("URGENT"), Value: "FATAL", Icon: ""},
				{Label: bdl.I18nPrinter.Sprintf("SERIOUS"), Value: "SERIOUS", Icon: ""},
				{Label: bdl.I18nPrinter.Sprintf("NORMAL"), Value: "NORMAL", Icon: ""},
				{Label: bdl.I18nPrinter.Sprintf("SLIGHT"), Value: "SLIGHT", Icon: ""},
				{Label: bdl.I18nPrinter.Sprintf("SUGGEST"), Value: "SUGGEST", Icon: ""},
			},
		},
		{
			Key:        PropConditionKeyCreatorIDs,
			Label:      bdl.I18nPrinter.Sprintf("Creator"),
			EmptyText:  bdl.I18nPrinter.Sprintf("All"),
			Fixed:      false,
			ShowIndex:  0,
			HaveFilter: true,
			Type:       filter.PropConditionTypeSelect,
			QuickSelect: filter.QuickSelect{
				Label:        bdl.I18nPrinter.Sprintf("Choose Yourself"),
				OperationKey: OperationKeyCreatorSelectMe,
			},
			Placeholder: "",
			Options:     nil,
		},
		{
			Key:        PropConditionKeyAssigneeIDs,
			Label:      bdl.I18nPrinter.Sprintf("Assignee"),
			EmptyText:  bdl.I18nPrinter.Sprintf("All"),
			Fixed:      false,
			ShowIndex:  0,
			HaveFilter: true,
			Type:       filter.PropConditionTypeSelect,
			QuickSelect: filter.QuickSelect{
				Label:        bdl.I18nPrinter.Sprintf("Choose Yourself"),
				OperationKey: OperationKeyAssigneeSelectMe,
			},
			Placeholder: "",
			Options:     nil,
		},
		{
			Key:        PropConditionKeyOwnerIDs,
			Label:      bdl.I18nPrinter.Sprintf("Responsible Person"),
			EmptyText:  bdl.I18nPrinter.Sprintf("All"),
			Fixed:      false,
			ShowIndex:  0,
			HaveFilter: true,
			Type:       filter.PropConditionTypeSelect,
			QuickSelect: filter.QuickSelect{
				Label:        bdl.I18nPrinter.Sprintf("Choose Yourself"),
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
					return bdl.I18nPrinter.Sprintf("Task Type")
				case apistructs.IssueTypeBug.String():
					return bdl.I18nPrinter.Sprintf("Import Source")
				}
				return ""
			}(),
			EmptyText:   bdl.I18nPrinter.Sprintf("All"),
			Fixed:       false,
			HaveFilter:  false,
			ShowIndex:   0,
			Type:        filter.PropConditionTypeSelect,
			QuickSelect: filter.QuickSelect{},
			Placeholder: "",
			Options: []filter.PropConditionOption{
				{Label: bdl.I18nPrinter.Sprintf("Demand Design"), Value: "demandDesign", Icon: ""},
				{Label: bdl.I18nPrinter.Sprintf("Architecture Design"), Value: "architectureDesign", Icon: ""},
				{Label: bdl.I18nPrinter.Sprintf("Code Development"), Value: "codeDevelopment", Icon: ""},
			},
		},
		{
			Key:         PropConditionKeyCreatedAtStartEnd,
			Label:       bdl.I18nPrinter.Sprintf("Created At"),
			EmptyText:   bdl.I18nPrinter.Sprintf("All"),
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
			Label:       bdl.I18nPrinter.Sprintf("Deadline"),
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
			Label:       bdl.I18nPrinter.Sprintf("Closed at"),
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
			Label:       bdl.I18nPrinter.Sprintf("State"),
			EmptyText:   bdl.I18nPrinter.Sprintf("All"),
			Fixed:       true,
			ShowIndex:   3,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: "",
			Options: func() []filter.PropConditionOption {
				open := filter.PropConditionOption{Label: bdl.I18nPrinter.Sprintf("OPEN"), Value: apistructs.IssueStateBelongOpen, Icon: ""}
				reopen := filter.PropConditionOption{Label: bdl.I18nPrinter.Sprintf("REOPEN"), Value: apistructs.IssueStateBelongReopen, Icon: ""}
				resolved := filter.PropConditionOption{Label: bdl.I18nPrinter.Sprintf("RESOLVED"), Value: apistructs.IssueStateBelongResloved, Icon: ""}
				wontfix := filter.PropConditionOption{Label: bdl.I18nPrinter.Sprintf("WONTFIX"), Value: apistructs.IssueStateBelongWontfix, Icon: ""}
				closed := filter.PropConditionOption{Label: bdl.I18nPrinter.Sprintf("CLOSED"), Value: apistructs.IssueStateBelongClosed, Icon: ""}
				working := filter.PropConditionOption{Label: bdl.I18nPrinter.Sprintf("WORKING"), Value: apistructs.IssueStateBelongWorking, Icon: ""}
				done := filter.PropConditionOption{Label: bdl.I18nPrinter.Sprintf("DONE"), Value: apistructs.IssueStateBelongDone, Icon: ""}
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
