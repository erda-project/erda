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
		PageSize:  300,
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
