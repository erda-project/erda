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
	model "github.com/erda-project/erda-infra/providers/component-protocol/components/filter/models"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
)

var (
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
)

func (f *IssueFilter) ConditionRetriever() ([]interface{}, error) {
	iterationOptions, err := f.getPropIterationsOptions()
	if err != nil {
		return nil, err
	}
	iterations := model.NewSelectCondition(PropConditionKeyIterationIDs, cputil.I18n(f.sdk.Ctx, "sprint"), iterationOptions).WithPlaceHolder(cputil.I18n(f.sdk.Ctx, "choose-sprint"))

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
	switch f.InParams.FrontendFixedIssueType {
	case apistructs.IssueTypeRequirement.String():
		conditions = []interface{}{iterations, labels, priority, creator, assignee, created, finished}
	case apistructs.IssueTypeTask.String():
		conditions = []interface{}{iterations, labels, priority, creator, assignee, stage, created, finished}
	case apistructs.IssueTypeBug.String():
		conditions = []interface{}{iterations, labels, priority, severity, creator, assignee, owner, stage, created, finished, closed}
	}
	return conditions, nil
	// stateBelongs := map[string][]apistructs.IssueStateBelong{
	// 	"TASK":        {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking},
	// 	"REQUIREMENT": {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking},
	// 	"BUG":         {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking, apistructs.IssueStateBelongWontfix, apistructs.IssueStateBelongReopen, apistructs.IssueStateBelongResolved},
	// 	"ALL":         {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking, apistructs.IssueStateBelongWontfix, apistructs.IssueStateBelongReopen, apistructs.IssueStateBelongResolved},
	// }[f.InParams.FrontendFixedIssueType]
	// types := []apistructs.IssueType{apistructs.IssueTypeRequirement, apistructs.IssueTypeTask, apistructs.IssueTypeBug}
	// res := make(map[string][]int64)
	// res["ALL"] = make([]int64, 0)
	// for _, v := range types {
	// 	req := &apistructs.IssueStatesGetRequest{
	// 		ProjectID:    f.InParams.ProjectID,
	// 		StateBelongs: stateBelongs,
	// 		IssueType:    v,
	// 	}
	// 	ids, err := f.issueStateSvc.GetIssueStateIDs(req)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	res[v.String()] = ids
	// 	res["ALL"] = append(res["ALL"], ids...)
	// }
}
