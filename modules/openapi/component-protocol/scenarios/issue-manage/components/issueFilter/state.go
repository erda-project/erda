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
	"fmt"

	"github.com/erda-project/erda/apistructs"
)

type State struct {
	// url 上带上参数，保证用户输入 url 时能定位过滤条件
	Base64UrlQueryParams string `json:"issueFilter__urlQuery,omitempty"`

	// 组件支持的过滤条件 state
	FrontendConditionProps  FrontendConditionProps `json:"conditions,omitempty"`
	FrontendConditionValues FrontendConditions     `json:"values,omitempty"`

	// 方便后端使用的 state
	IssuePagingRequest apistructs.IssuePagingRequest `json:"issuePagingRequest,omitempty"`

	IssueViewGroupValue         string            `json:"issueViewGroupValue"`
	IssueViewGroupChildrenValue map[string]string `json:"issueViewGroupChildrenValue"`
}

// generateUrlQueryKey 实际上组件名在一个协议里是定义好的，即 issueFilter__urlQuery
func (f *ComponentFilter) generateUrlQueryKey() string {
	return fmt.Sprintf("%s__urlQuery", f.Name)
}

func (f *ComponentFilter) generateIssuePagingRequest() (apistructs.IssuePagingRequest, error) {
	var (
		startCreatedAt, endCreatedAt, startFinishedAt, endFinishedAt, startClosedAt, endClosedAt int64
	)
	if len(f.State.FrontendConditionValues.CreatedAtStartEnd) >= 2 {
		if f.State.FrontendConditionValues.CreatedAtStartEnd[0] != nil {
			startCreatedAt = *f.State.FrontendConditionValues.CreatedAtStartEnd[0]
			if f.State.FrontendConditionValues.CreatedAtStartEnd[1] == nil {
				endCreatedAt = 0
			} else {
				endCreatedAt = *f.State.FrontendConditionValues.CreatedAtStartEnd[1]
			}
		} else if f.State.FrontendConditionValues.CreatedAtStartEnd[1] != nil {
			startCreatedAt = 0
			endCreatedAt = *f.State.FrontendConditionValues.CreatedAtStartEnd[1]
		}
	}
	if len(f.State.FrontendConditionValues.FinishedAtStartEnd) >= 2 {
		if f.State.FrontendConditionValues.FinishedAtStartEnd[0] != nil {
			startFinishedAt = *f.State.FrontendConditionValues.FinishedAtStartEnd[0]
			if f.State.FrontendConditionValues.FinishedAtStartEnd[1] == nil {
				endFinishedAt = 0
			} else {
				endFinishedAt = *f.State.FrontendConditionValues.FinishedAtStartEnd[1]
			}
		} else if f.State.FrontendConditionValues.FinishedAtStartEnd[1] != nil {
			startFinishedAt = 0
			endFinishedAt = *f.State.FrontendConditionValues.FinishedAtStartEnd[1]
		}
	}
	if len(f.State.FrontendConditionValues.ClosedAtStartEnd) >= 2 {
		if f.State.FrontendConditionValues.ClosedAtStartEnd[0] != nil {
			startClosedAt = *f.State.FrontendConditionValues.ClosedAtStartEnd[0]
			if f.State.FrontendConditionValues.ClosedAtStartEnd[1] == nil {
				endClosedAt = 0
			} else {
				endClosedAt = *f.State.FrontendConditionValues.ClosedAtStartEnd[1]
			}
		} else if f.State.FrontendConditionValues.ClosedAtStartEnd[1] != nil {
			startClosedAt = 0
			endClosedAt = *f.State.FrontendConditionValues.ClosedAtStartEnd[1]
		}
	}

	req := apistructs.IssuePagingRequest{
		PageNo:   1, // 每次走 filter，都需要重新查询，调整 pageNo 为 1
		PageSize: 0,
		OrgID:    int64(f.InParams.OrgID),
		IssueListRequest: apistructs.IssueListRequest{
			Title:           f.State.FrontendConditionValues.Title,
			Type:            f.InParams.IssueTypes,
			ProjectID:       f.InParams.ProjectID,
			IterationID:     f.InParams.IterationID,
			IterationIDs:    f.State.FrontendConditionValues.IterationIDs,
			AppID:           nil,
			RequirementID:   nil,
			State:           nil,
			StateBelongs:    f.State.FrontendConditionValues.StateBelongs,
			Creators:        f.State.FrontendConditionValues.CreatorIDs,
			Assignees:       f.State.FrontendConditionValues.AssigneeIDs,
			Label:           f.State.FrontendConditionValues.LabelIDs,
			StartCreatedAt:  startCreatedAt,
			EndCreatedAt:    endCreatedAt,
			StartFinishedAt: startFinishedAt,
			EndFinishedAt:   endFinishedAt,
			StartClosedAt:   startClosedAt,
			EndClosedAt:     endClosedAt,
			Priority:        f.State.FrontendConditionValues.Priorities,
			Complexity:      nil,
			Severity:        f.State.FrontendConditionValues.Severities,
			RelatedIssueIDs: nil,
			Source:          "",
			OrderBy:         "",
			TaskType:        nil,
			BugStage:        f.State.FrontendConditionValues.BugStages,
			Owner:           f.State.FrontendConditionValues.OwnerIDs,
			Asc:             false,
			IDs:             nil,
			IdentityInfo:    apistructs.IdentityInfo{UserID: f.CtxBdl.Identity.UserID},
			External:        false,

			WithProcessSummary: f.InParams.FrontendFixedIssueType == apistructs.IssueTypeRequirement.String(),
		},
	}
	return req, nil
}
