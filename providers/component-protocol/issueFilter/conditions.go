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

	model "github.com/erda-project/erda-infra/providers/component-protocol/components/filter/models"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/providers/component-protocol/condition"
)

const (
	PropConditionKeyFilterID           string = "filterID" // special, need emit it when hashing filter
	PropConditionKeyIterationIDs       string = "iterationIDs"
	PropConditionKeyTitle              string = "title"
	PropConditionKeyStateBelongs       string = "stateBelongs"
	PropConditionKeyStates             string = "states"
	PropConditionKeyLabelIDs           string = "labelIDs"
	PropConditionKeyPriorities         string = "priorities"
	PropConditionKeySeverities         string = "severities"
	PropConditionKeyCreatorIDs         string = "creatorIDs"
	PropConditionKeyAssigneeIDs        string = "assigneeIDs"
	PropConditionKeyOwnerIDs           string = "ownerIDs"
	PropConditionKeyBugStages          string = "bugStages"
	PropConditionKeyCreatedAtStartEnd  string = "createdAtStartEnd"
	PropConditionKeyFinishedAtStartEnd string = "finishedAtStartEnd"
	PropConditionKeyClosed             string = "closedAtStartEnd"
	PropConditionKeyComplexity         string = "complexities"
)

type FrontendConditions struct {
	FilterID           string                        `json:"filterID,omitempty"`
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
	Complexities       []apistructs.IssueComplexity  `json:"complexities,omitempty"`
}

func (f *IssueFilter) ConditionRetriever() ([]interface{}, error) {
	needIterationCond := true
	if f.InParams.FrontendFixedIteration != "" {
		needIterationCond = false
	}
	var iterations *model.SelectCondition
	if needIterationCond {
		iterationOptions, err := f.getPropIterationsOptions()
		if err != nil {
			return nil, err
		}
		iterations = model.NewSelectCondition(PropConditionKeyIterationIDs, cputil.I18n(f.sdk.Ctx, "sprint"), iterationOptions).WithPlaceHolder(cputil.I18n(f.sdk.Ctx, "choose-sprint"))
	}

	labelOptions, err := f.getPropLabelsOptions()
	if err != nil {
		return nil, err
	}
	labels := model.NewTagsSelectCondition(PropConditionKeyLabelIDs, cputil.I18n(f.sdk.Ctx, "label"), labelOptions)
	labels.ConditionBase.Placeholder = cputil.I18n(f.sdk.Ctx, "please-choose-label")

	priorityOptions := []model.SelectOption{
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "urgent"), "URGENT"),
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "high"), "HIGH"),
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "normal"), "NORMAL"),
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "low"), "LOW"),
	}
	priority := model.NewSelectCondition(PropConditionKeyPriorities, cputil.I18n(f.sdk.Ctx, "priority"), priorityOptions).WithPlaceHolder(cputil.I18n(f.sdk.Ctx, "choose-priorities"))

	severityOptions := []model.SelectOption{
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "fatal"), "FATAL"),
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "serious"), "SERIOUS"),
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "ordinary"), "NORMAL"),
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "slight"), "SLIGHT"),
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "suggest"), "SUGGEST"),
	}
	severity := model.NewSelectCondition(PropConditionKeySeverities, cputil.I18n(f.sdk.Ctx, "severity"), severityOptions).WithPlaceHolder(cputil.I18n(f.sdk.Ctx, "choose-severity"))

	memberOptions, err := f.getProjectMemberOptions()
	if err != nil {
		return nil, err
	}
	creator := model.NewSelectCondition(PropConditionKeyCreatorIDs, cputil.I18n(f.sdk.Ctx, "creator"), memberOptions)
	assignee := model.NewSelectCondition(PropConditionKeyAssigneeIDs, cputil.I18n(f.sdk.Ctx, "assignee"), memberOptions)
	owner := model.NewSelectCondition(PropConditionKeyOwnerIDs, cputil.I18n(f.sdk.Ctx, "responsible-person"), memberOptions)

	stageOptions := []model.SelectOption{
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "demand-design"), "demandDesign"),
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "architecture-design"), "architectureDesign"),
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "code-development"), "codeDevelopment"),
	}
	if f.InParams.FrontendFixedIssueType == apistructs.IssueTypeTask.String() || f.InParams.FrontendFixedIssueType == apistructs.IssueTypeBug.String() {
		stageOptions, err = f.getPropStagesOptions(f.InParams.FrontendFixedIssueType)
		if err != nil {
			return nil, err
		}
	}
	stage := model.NewSelectCondition(PropConditionKeyBugStages, func() string {
		switch f.InParams.FrontendFixedIssueType {
		case "ALL":
		case apistructs.IssueTypeEpic.String():
		case apistructs.IssueTypeRequirement.String():
		case apistructs.IssueTypeTask.String():
			return cputil.I18n(f.sdk.Ctx, "task-type")
		case apistructs.IssueTypeBug.String():
			return cputil.I18n(f.sdk.Ctx, "import-source")
		}
		return ""
	}(), stageOptions)

	created := model.NewDateRangeCondition(PropConditionKeyCreatedAtStartEnd, cputil.I18n(f.sdk.Ctx, "created-at"))
	finished := model.NewDateRangeCondition(PropConditionKeyFinishedAtStartEnd, cputil.I18n(f.sdk.Ctx, "deadline"))
	closed := model.NewDateRangeCondition(PropConditionKeyClosed, cputil.I18n(f.sdk.Ctx, "closed-at"))

	var conditions []interface{}
	if needIterationCond {
		conditions = []interface{}{iterations}
	}

	complexityOptions := []model.SelectOption{
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "HARD"), "HARD"),
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "NORMAL"), "NORMAL"),
		*model.NewSelectOption(cputil.I18n(f.sdk.Ctx, "EASY"), "EASY"),
	}
	complexity := model.NewSelectCondition(PropConditionKeyComplexity, cputil.I18n(f.sdk.Ctx, "complexity"), complexityOptions)

	// statesMap, err := f.issueStateSvc.GetIssueStatesMap(&apistructs.IssueStatesGetRequest{
	// 	ProjectID: f.InParams.ProjectID,
	// })
	// if err != nil {
	// 	return nil, err
	// }

	// status := func() interface{} {
	// 	switch f.InParams.FrontendFixedIssueType {
	// 	case "ALL":
	// 		return model.NewSelectConditionWithChildren(PropConditionKeyStates, cputil.I18n(f.sdk.Ctx, "state"), convertAllConditions(f.sdk.Ctx, statesMap))
	// 	case apistructs.IssueTypeRequirement.String():
	// 		return model.NewSelectCondition(PropConditionKeyStates, cputil.I18n(f.sdk.Ctx, "state"), convertConditions(statesMap[apistructs.IssueTypeRequirement]))
	// 	case apistructs.IssueTypeTask.String():
	// 		return model.NewSelectCondition(PropConditionKeyStates, cputil.I18n(f.sdk.Ctx, "state"), convertConditions(statesMap[apistructs.IssueTypeTask]))
	// 	case apistructs.IssueTypeBug.String():
	// 		return model.NewSelectCondition(PropConditionKeyStates, cputil.I18n(f.sdk.Ctx, "state"), convertConditions(statesMap[apistructs.IssueTypeBug]))
	// 	}
	// 	return nil
	// }()

	switch f.InParams.FrontendFixedIssueType {
	case apistructs.IssueTypeRequirement.String():
		conditions = append(conditions, labels, priority, complexity, creator, assignee, created, finished)
	case apistructs.IssueTypeTask.String():
		conditions = append(conditions, labels, priority, complexity, creator, assignee, stage, created, finished)
	case apistructs.IssueTypeBug.String():
		conditions = append(conditions, labels, priority, complexity, severity, creator, assignee, owner, stage, created, finished, closed)
	case "ALL":
		conditions = append(conditions, labels, priority, complexity, creator, assignee, owner, created, finished)
	}

	conditions = append(conditions, condition.ExternalInputCondition("title", "title", cputil.I18n(f.sdk.Ctx, "searchByName")))
	return conditions, nil
}

func convertConditions(status []apistructs.IssueStatus) []model.SelectOption {
	options := make([]model.SelectOption, 0, len(status))
	for _, i := range status {
		options = append(options, *model.NewSelectOption(i.StateName, i.StateID))
	}
	return options
}

func convertAllConditions(ctx context.Context, stateMap map[apistructs.IssueType][]apistructs.IssueStatus) []model.SelectOptionWithChildren {
	options := make([]model.SelectOptionWithChildren, 0, len(stateMap))
	for i, v := range stateMap {
		options = append(options, model.SelectOptionWithChildren{
			SelectOption: *model.NewSelectOption(cputil.I18n(ctx, strings.ToLower(i.String())), i.String()),
			Children:     convertConditions(v),
		})
	}
	return options
}
