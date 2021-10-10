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

package filter

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/services/issue"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

func init() {
	base.InitProviderWithCreator("issue-dashboard", "filter",
		func() servicehub.Provider { return &ComponentFilter{} })
}

func (f *ComponentFilter) InitFromProtocol(ctx context.Context, c *cptype.Component) error {
	// component 序列化
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, f); err != nil {
		return err
	}

	// sdk
	f.sdk = cputil.SDK(ctx)
	f.bdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	f.issueSvc = ctx.Value(types.IssueService).(*issue.Issue)
	if err := f.setInParams(ctx); err != nil {
		return err
	}

	return nil
}

func (f *ComponentFilter) setInParams(ctx context.Context) error {
	b, err := json.Marshal(cputil.SDK(ctx).InParams)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &f.InParams); err != nil {
		return err
	}

	return nil
}

// func (f *ComponentFilter) GenComponentState(c *cptype.Component) error {
// 	if c == nil || c.State == nil {
// 		return nil
// 	}
// 	var state State
// 	cont, err := json.Marshal(c.State)
// 	if err != nil {
// 		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
// 		return err
// 	}
// 	err = json.Unmarshal(cont, &state)
// 	if err != nil {
// 		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
// 		return err
// 	}
// 	f.State = state
// 	return nil
// }

func (f *ComponentFilter) SetToProtocolComponent(c *cptype.Component) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}
	return nil
}

func (f *ComponentFilter) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := f.InitFromProtocol(ctx, c); err != nil {
		return err
	}
	// if err := f.GenComponentState(c); err != nil {
	// 	return err
	// }

	switch event.Operation {
	case cptype.InitializeOperation, cptype.RenderingOperation:
		if err := f.InitDefaultOperation(ctx, f.State); err != nil {
			return err
		}
	case cptype.OperationKey(f.Operations[OperationKeyFilter].Key):
	}

	data, err := f.issueSvc.GetAllIssuesByProject(apistructs.IssueListRequest{
		Type: []apistructs.IssueType{
			apistructs.IssueTypeBug,
		},
		ProjectID:    f.InParams.ProjectID,
		IterationIDs: f.State.Values.IterationIDs,
		Assignees:    f.State.Values.AssigneeIDs,
		// StateBelongs: []apistructs.IssueStateBelong{apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking, apistructs.IssueStateBelongWontfix, apistructs.IssueStateBelongReopen, apistructs.IssueStateBelongResloved},
	})
	if err != nil {
		return err
	}

	// todo modify data format
	f.State.IssueList = data

	states, err := f.issueSvc.GetIssuesStatesByProjectID(f.InParams.ProjectID, apistructs.IssueTypeBug)
	if err != nil {
		return err
	}
	f.State.IssueStateList = states

	urlParam, err := f.generateUrlQueryParams()
	if err != nil {
		return err
	}
	f.State.Base64UrlQueryParams = urlParam
	return f.SetToProtocolComponent(c)
}

func (f *ComponentFilter) InitDefaultOperation(ctx context.Context, state State) error {
	iterations, iterationOptions, err := f.getPropIterationsOptions()
	if err != nil {
		return err
	}

	projectMemberOptions, err := f.getProjectMemberOptions()
	if err != nil {
		return err
	}

	f.State.Conditions = []filter.PropCondition{
		{
			EmptyText: "全部",
			Fixed:     true,
			Key:       "iteration",
			Label:     "迭代",
			Options:   iterationOptions,
			Type:      filter.PropConditionTypeSelect,
		},
		{
			EmptyText: "全部",
			Fixed:     true,
			Key:       "member",
			Label:     "成员",
			Options:   projectMemberOptions,
			Type:      filter.PropConditionTypeSelect,
		},
	}

	f.Props = filter.Props{
		Delay: 1000,
	}
	f.Operations = map[filter.OperationKey]filter.Operation{
		OperationKeyFilter: {
			Key:    OperationKeyFilter,
			Reload: true,
		},
	}
	if f.State.Base64UrlQueryParams != "" {

	} else { // todo default value
		f.State.Values.IterationIDs = []int64{iterationOptions[0].Value.(int64)}
		// f.State.Values.AssigneeIDs = []string{"2"}
	}

	// TODO select iteration
	f.State.Iterations = iterations
	return nil
}

func (f *ComponentFilter) generateUrlQueryParams() (string, error) {
	fb, err := json.Marshal(f.State.Values)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(fb), nil
}

func (f *ComponentFilter) getPropIterationsOptions() ([]apistructs.Iteration, []filter.PropConditionOption, error) {
	f.sdk.Identity.OrgID = "1"
	iterations, err := f.bdl.ListProjectIterations(apistructs.IterationPagingRequest{PageNo: 1, PageSize: 1000, ProjectID: f.InParams.ProjectID, WithoutIssueSummary: true}, f.sdk.Identity.OrgID)
	if err != nil {
		return nil, nil, err
	}
	var options []filter.PropConditionOption
	for _, iteration := range iterations {
		options = append(options, filter.PropConditionOption{
			Label: iteration.Title,
			Value: iteration.ID,
		})
	}
	return iterations, options, nil
}

func (f *ComponentFilter) getProjectMemberOptions() ([]filter.PropConditionOption, error) {
	members, err := f.bdl.ListMembers(apistructs.MemberListRequest{
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
		})
	}
	return results, nil
}
