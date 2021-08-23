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
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

func (f *ComponentFilter) getPropIterationsOptions() ([]filter.PropConditionOption, error) {
	iterations, err := f.CtxBdl.Bdl.ListProjectIterations(apistructs.IterationPagingRequest{PageNo: 1, PageSize: 1000, ProjectID: f.InParams.ProjectID, WithoutIssueSummary: true}, f.CtxBdl.Identity.OrgID)
	if err != nil {
		return nil, err
	}
	var options []filter.PropConditionOption
	for _, iteration := range iterations {
		options = append(options, filter.PropConditionOption{
			Label: iteration.Title,
			Value: iteration.ID,
			Icon:  "",
		})
	}
	return options, nil
}

func (f *ComponentFilter) getPropLabelsOptions() ([]filter.PropConditionOption, error) {
	labels, err := f.CtxBdl.Bdl.Labels(string(apistructs.LabelTypeIssue), f.InParams.ProjectID, f.CtxBdl.Identity.UserID)
	if err != nil {
		return nil, err
	}
	if labels == nil {
		return nil, nil
	}
	var options []filter.PropConditionOption
	for _, label := range labels.List {
		options = append(options, filter.PropConditionOption{
			Label: label.Name,
			Value: label.ID,
			Icon:  "",
		})
	}
	return options, nil
}

func (f *ComponentFilter) getProjectMemberOptions() ([]filter.PropConditionOption, error) {
	members, err := f.CtxBdl.Bdl.ListMembers(apistructs.MemberListRequest{
		ScopeType: apistructs.ProjectScope,
		ScopeID:   int64(f.InParams.ProjectID),
		PageNo:    1,
		PageSize:  500,
	})
	if err != nil {
		return nil, err
	}
	var results []filter.PropConditionOption
	for _, member := range members {
		results = append(results, filter.PropConditionOption{
			Label: func() string {
				if member.Nick != "" {
					return member.Nick
				}
				if member.Name != "" {
					return member.Name
				}
				return member.Mobile
			}(),
			Value: member.UserID,
			Icon:  "",
		})
	}
	return results, nil
}

func (f *ComponentFilter) getPropStagesOptions(tp string) ([]filter.PropConditionOption, error) {
	stages, err := f.CtxBdl.Bdl.GetIssueStage(int64(f.InParams.OrgID), apistructs.IssueType(tp))
	if err != nil {
		return nil, err
	}
	var options []filter.PropConditionOption
	for _, stage := range stages {
		options = append(options, filter.PropConditionOption{
			Label: stage.Name,
			Value: stage.Value,
			Icon:  "",
		})
	}
	return options, nil
}
