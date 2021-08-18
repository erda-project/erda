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
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
	"github.com/erda-project/erda/pkg/strutil"
)

type FrontendConditionProps []filter.PropCondition

func (f *ComponentFilter) SetStateConditionProps() ([]filter.PropCondition, error) {
	var err error
	var onceProjectMemberOptions []filter.PropConditionOption
	var onceProjectMemberQueried bool

	var newConditions []filter.PropCondition

	for i := range f.State.FrontendConditionProps {
		cond := f.State.FrontendConditionProps[i]
		flag := true
		switch cond.Key {
		case PropConditionKeyIterationIDs:
			cond.Options, err = f.getPropIterationsOptions()
			if err != nil {
				return nil, err
			}
			// 从专门的迭代中跳转进来，过滤项中删除该字段
			if f.InParams.IterationID > 0 {
				flag = false
				break
			}

		case PropConditionKeyTitle:

		case PropConditionKeyStateBelongs:

		case PropConditionKeyLabelIDs:
			cond.Options, err = f.getPropLabelsOptions()
			if err != nil {
				return nil, err
			}

		case PropConditionKeyPriorities:

		case PropConditionKeySeverities:

		case PropConditionKeyCreatorIDs:
			if !onceProjectMemberQueried {
				projectMemberOptions, err := f.getProjectMemberOptions()
				if err != nil {
					return nil, err
				}
				onceProjectMemberOptions = projectMemberOptions
				onceProjectMemberQueried = true
			}
			cond.Options = onceProjectMemberOptions
			if userIDs := f.State.FrontendConditionValues.CreatorIDs; userIDs != nil {
				cond.Options = reorderMemberOption(onceProjectMemberOptions, userIDs)
			}

		case PropConditionKeyAssigneeIDs:
			if !onceProjectMemberQueried {
				projectMemberOptions, err := f.getProjectMemberOptions()
				if err != nil {
					return nil, err
				}
				onceProjectMemberOptions = projectMemberOptions
				onceProjectMemberQueried = true
			}
			cond.Options = onceProjectMemberOptions
			if userIDs := f.State.FrontendConditionValues.AssigneeIDs; userIDs != nil {
				cond.Options = reorderMemberOption(onceProjectMemberOptions, userIDs)
			}

		case PropConditionKeyOwnerIDs:
			if f.InParams.FrontendFixedIssueType == "TASK" || f.InParams.FrontendFixedIssueType == "REQUIREMENT" {
				flag = false
				break
			}
			if !onceProjectMemberQueried {
				projectMemberOptions, err := f.getProjectMemberOptions()
				if err != nil {
					return nil, err
				}
				onceProjectMemberOptions = projectMemberOptions
				onceProjectMemberQueried = true
			}
			cond.Options = onceProjectMemberOptions
			if userIDs := f.State.FrontendConditionValues.OwnerIDs; userIDs != nil {
				cond.Options = reorderMemberOption(onceProjectMemberOptions, userIDs)
			}

		case PropConditionKeyBugStages:
			if f.InParams.FrontendFixedIssueType == "REQUIREMENT" || f.InParams.FrontendFixedIssueType == "ALL" {
				flag = false
				break
			}
			cond.Options, err = f.getPropStagesOptions(f.InParams.FrontendFixedIssueType)
			if err != nil {
				return nil, err
			}

		case PropConditionKeyCreatedAtStartEnd:

		case PropConditionKeyFinishedAtStartEnd:

		case PropConditionKeyClosed:

		}
		if flag {
			newConditions = append(newConditions, cond)
		}
	}

	return newConditions, nil
}

func reorderMemberOption(options []filter.PropConditionOption, userIDs []string) []filter.PropConditionOption {
	var selected []filter.PropConditionOption
	var result []filter.PropConditionOption
	for _, option := range options {
		if strutil.Exist(userIDs, option.Value.(string)) {
			selected = append(selected, option)
		} else {
			result = append(result, option)
		}
	}
	return append(selected, result...)
}
