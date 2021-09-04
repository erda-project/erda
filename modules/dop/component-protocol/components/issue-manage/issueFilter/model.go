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
	"context"
	"strings"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/services/issuestate"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

type ComponentFilter struct {
	sdk           *cptype.SDK
	bdl           *bundle.Bundle
	issueStateSvc *issuestate.IssueState
	filter.CommonFilter
	State    State    `json:"state,omitempty"`
	InParams InParams `json:"-"`
	base.DefaultProvider
}

// FrontendConditions 前端支持的过滤参数
type FrontendConditions struct {
	IterationIDs       []int64                       `json:"iterationIDs,omitempty"`
	Title              string                        `json:"title,omitempty"`
	StateBelongs       []apistructs.IssueStateBelong `json:"stateBelongs,omitempty"`
	States             []int64                       `json:"states,omitempty"`
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

func (f *ComponentFilter) generateFrontendConditionProps(ctx context.Context, fixedIssueType string, state State) FrontendConditionProps {
	conditionProps := []filter.PropCondition{
		{
			Key:         PropConditionKeyIterationIDs,
			Label:       cputil.I18n(ctx, "sprint"),
			EmptyText:   cputil.I18n(ctx, "all"),
			Fixed:       true,
			ShowIndex:   1,
			HaveFilter:  true,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: cputil.I18n(ctx, "choose-sprint"),
			Options:     nil,
		},
		{
			Key:         PropConditionKeyTitle,
			Label:       cputil.I18n(ctx, "title"),
			EmptyText:   cputil.I18n(ctx, "all"),
			Fixed:       true,
			ShowIndex:   2,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeInput,
			Placeholder: cputil.I18n(ctx, "please-enter-title-or-id"),
		},
		{
			Key:         PropConditionKeyLabelIDs,
			Label:       cputil.I18n(ctx, "label"),
			EmptyText:   cputil.I18n(ctx, "all"),
			Fixed:       false,
			ShowIndex:   0,
			HaveFilter:  true,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: cputil.I18n(ctx, "please-choose-label"),
			Options:     nil,
		},
		{
			Key:         PropConditionKeyPriorities,
			Label:       cputil.I18n(ctx, "priority"),
			EmptyText:   cputil.I18n(ctx, "all"),
			Fixed:       false,
			ShowIndex:   0,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: cputil.I18n(ctx, "choose-priorities"),
			Options: []filter.PropConditionOption{
				{Label: cputil.I18n(ctx, "urgent"), Value: "URGENT", Icon: ""},
				{Label: cputil.I18n(ctx, "high"), Value: "HIGH", Icon: ""},
				{Label: cputil.I18n(ctx, "normal"), Value: "NORMAL", Icon: ""},
				{Label: cputil.I18n(ctx, "low"), Value: "LOW", Icon: ""},
			},
		},
		{
			Key:         PropConditionKeySeverities,
			Label:       cputil.I18n(ctx, "severity"),
			EmptyText:   cputil.I18n(ctx, "all"),
			Fixed:       false,
			ShowIndex:   0,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: cputil.I18n(ctx, "choose-severity"),
			Options: []filter.PropConditionOption{
				{Label: cputil.I18n(ctx, "fatal"), Value: "FATAL", Icon: ""},
				{Label: cputil.I18n(ctx, "serious"), Value: "SERIOUS", Icon: ""},
				{Label: cputil.I18n(ctx, "normal"), Value: "NORMAL", Icon: ""},
				{Label: cputil.I18n(ctx, "slight"), Value: "SLIGHT", Icon: ""},
				{Label: cputil.I18n(ctx, "suggest"), Value: "SUGGEST", Icon: ""},
			},
		},
		{
			Key:        PropConditionKeyCreatorIDs,
			Label:      cputil.I18n(ctx, "creator"),
			EmptyText:  cputil.I18n(ctx, "all"),
			Fixed:      false,
			ShowIndex:  0,
			HaveFilter: true,
			Type:       filter.PropConditionTypeSelect,
			QuickSelect: filter.QuickSelect{
				Label:        cputil.I18n(ctx, "choose-yourself"),
				OperationKey: OperationKeyCreatorSelectMe,
			},
			Placeholder: "",
			Options:     nil,
		},
		{
			Key:        PropConditionKeyAssigneeIDs,
			Label:      cputil.I18n(ctx, "assignee"),
			EmptyText:  cputil.I18n(ctx, "all"),
			Fixed:      false,
			ShowIndex:  0,
			HaveFilter: true,
			Type:       filter.PropConditionTypeSelect,
			QuickSelect: filter.QuickSelect{
				Label:        cputil.I18n(ctx, "choose-yourself"),
				OperationKey: OperationKeyAssigneeSelectMe,
			},
			Placeholder: "",
			Options:     nil,
		},
		{
			Key:        PropConditionKeyOwnerIDs,
			Label:      cputil.I18n(ctx, "responsible-person"),
			EmptyText:  cputil.I18n(ctx, "all"),
			Fixed:      false,
			ShowIndex:  0,
			HaveFilter: true,
			Type:       filter.PropConditionTypeSelect,
			QuickSelect: filter.QuickSelect{
				Label:        cputil.I18n(ctx, "choose-yourself"),
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
					return cputil.I18n(ctx, "task-type")
				case apistructs.IssueTypeBug.String():
					return cputil.I18n(ctx, "import-source")
				}
				return ""
			}(),
			EmptyText:   cputil.I18n(ctx, "all"),
			Fixed:       false,
			HaveFilter:  false,
			ShowIndex:   0,
			Type:        filter.PropConditionTypeSelect,
			QuickSelect: filter.QuickSelect{},
			Placeholder: "",
			Options: []filter.PropConditionOption{
				{Label: cputil.I18n(ctx, "demand-design"), Value: "demandDesign", Icon: ""},
				{Label: cputil.I18n(ctx, "architecture-design"), Value: "architectureDesign", Icon: ""},
				{Label: cputil.I18n(ctx, "code-development"), Value: "codeDevelopment", Icon: ""},
			},
		},
		{
			Key:         PropConditionKeyCreatedAtStartEnd,
			Label:       cputil.I18n(ctx, "created-at"),
			EmptyText:   cputil.I18n(ctx, "all"),
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
			Label:       cputil.I18n(ctx, "deadline"),
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
			Label:       cputil.I18n(ctx, "closed-at"),
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
		statesMap, err := f.issueStateSvc.GetIssueStatesMap(&apistructs.IssueStatesGetRequest{
			ProjectID: f.InParams.ProjectID,
		})
		if err != nil {
			return nil
		}

		status := filter.PropCondition{
			Key:         PropConditionKeyStates,
			Label:       cputil.I18n(ctx, "state"),
			EmptyText:   cputil.I18n(ctx, "all"),
			Fixed:       true,
			ShowIndex:   3,
			HaveFilter:  false,
			Type:        filter.PropConditionTypeSelect,
			Placeholder: "",
			Options: func() []filter.PropConditionOption {
				// open := filter.PropConditionOption{Label: cputil.I18n(ctx, "open"), Value: apistructs.IssueStateBelongOpen, Icon: ""}
				// reopen := filter.PropConditionOption{Label: cputil.I18n(ctx, "reopen"), Value: apistructs.IssueStateBelongReopen, Icon: ""}
				// resolved := filter.PropConditionOption{Label: cputil.I18n(ctx, "resolved"), Value: apistructs.IssueStateBelongResloved, Icon: ""}
				// wontfix := filter.PropConditionOption{Label: cputil.I18n(ctx, "wontfix"), Value: apistructs.IssueStateBelongWontfix, Icon: ""}
				// closed := filter.PropConditionOption{Label: cputil.I18n(ctx, "closed"), Value: apistructs.IssueStateBelongClosed, Icon: ""}
				// working := filter.PropConditionOption{Label: cputil.I18n(ctx, "working"), Value: apistructs.IssueStateBelongWorking, Icon: ""}
				// done := filter.PropConditionOption{Label: cputil.I18n(ctx, "done"), Value: apistructs.IssueStateBelongDone, Icon: ""}
				switch fixedIssueType {
				case "ALL":
					return convertAllConditions(ctx, statesMap)
				case apistructs.IssueTypeRequirement.String():
					return convertConditions(statesMap[apistructs.IssueTypeRequirement])
				case apistructs.IssueTypeTask.String():
					return convertConditions(statesMap[apistructs.IssueTypeTask])
				case apistructs.IssueTypeBug.String():
					return convertConditions(statesMap[apistructs.IssueTypeBug])
				}
				return nil
			}(),
		}
		conditionProps = append(conditionProps[:2], append([]filter.PropCondition{status}, conditionProps[2:]...)...)
	}

	return conditionProps
}

func convertConditions(status []apistructs.IssueStatus) []filter.PropConditionOption {
	options := make([]filter.PropConditionOption, 0, len(status))
	for _, i := range status {
		options = append(options, filter.PropConditionOption{
			Label: i.StateName,
			Value: i.StateID,
		})
	}
	return options
}

func convertAllConditions(ctx context.Context, stateMap map[apistructs.IssueType][]apistructs.IssueStatus) []filter.PropConditionOption {
	options := make([]filter.PropConditionOption, 0, len(stateMap))
	for i, v := range stateMap {
		child := convertConditions(v)
		opt := filter.PropConditionOption{
			Label:    cputil.I18n(ctx, strings.ToLower(i.String())),
			Value:    i.String(),
			Children: child,
		}
		options = append(options, opt)
	}
	return options
}

var (
	PropConditionKeyIterationIDs       filter.PropConditionKey = "iterationIDs"
	PropConditionKeyTitle              filter.PropConditionKey = "title"
	PropConditionKeyStateBelongs       filter.PropConditionKey = "stateBelongs"
	PropConditionKeyStates             filter.PropConditionKey = "states"
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
