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
	"fmt"
	"strconv"
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/services/issue"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
	"github.com/erda-project/erda/pkg/strutil"
)

func init() {
	base.InitProviderWithCreator("issue-gantt", "filter",
		func() servicehub.Provider { return &ComponentFilter{} })
}

func (f *ComponentFilter) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	f.sdk = cputil.SDK(ctx)
	f.bdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	f.issueSvc = ctx.Value(types.IssueService).(*issue.Issue)
	projectID, err := strconv.ParseUint(cputil.GetInParamByKey(ctx, "projectId").(string), 10, 64)
	if err != nil {
		return err
	}
	f.projectID = projectID
	if fixedIterationIDStr := strutil.String(cputil.GetInParamByKey(ctx, "fixedIteration")); fixedIterationIDStr != "" {
		fixedIterationID, err := strconv.ParseUint(fixedIterationIDStr, 10, 64)
		if err != nil {
			return err
		}
		f.fixedIterationID = fixedIterationID
	}
	if q := cputil.GetInParamByKey(ctx, "filter__urlQuery"); q != nil {
		f.FrontendUrlQuery = q.(string)
	}

	iteration, iterationOptions, err := f.getPropIterationsOptions()
	if err != nil {
		return err
	}
	projectMemberOptions, err := f.getProjectMemberOptions()
	if err != nil {
		return err
	}
	labelOptions, err := f.getProjectLabelsOptions()
	if err != nil {
		return err
	}

	switch event.Operation {
	case cptype.InitializeOperation, cptype.RenderingOperation:
		f.Props = filter.Props{
			Delay: 1000,
		}
		f.Operations = map[filter.OperationKey]filter.Operation{
			OperationKeyFilter: {
				Key:    OperationKeyFilter,
				Reload: true,
			},
		}
		if f.FrontendUrlQuery != "" {
			b, err := base64.StdEncoding.DecodeString(f.FrontendUrlQuery)
			if err != nil {
				return err
			}
			f.State.Values = FrontendConditions{}
			if err := json.Unmarshal(b, &f.State.Values); err != nil {
				return err
			}
		} else {
			f.State.Values.IterationIDs = []int64{iteration}
		}
		if f.fixedIterationID > 0 {
			f.State.Values.IterationIDs = []int64{iteration}
		}
	}

	var propConditions []filter.PropCondition
	emptyText := cputil.I18n(ctx, "all")
	// only show iteration prop when fixedIterationID not exist
	if f.fixedIterationID == 0 {
		propConditions = append(propConditions, filter.PropCondition{
			EmptyText:  emptyText,
			Fixed:      true,
			Key:        "iteration",
			Label:      cputil.I18n(ctx, "iteration"),
			Options:    iterationOptions,
			Type:       filter.PropConditionTypeSelect,
			HaveFilter: true,
			// Required:   true,
			// CustomProps: map[string]interface{}{
			// 	"mode": "single",
			// },
		})
	}
	propConditions = append(propConditions,
		filter.PropCondition{
			EmptyText:  emptyText,
			Fixed:      true,
			Key:        "member",
			Label:      cputil.I18n(ctx, "member"),
			Options:    projectMemberOptions,
			Type:       filter.PropConditionTypeSelect,
			HaveFilter: true,
		},
		filter.PropCondition{
			EmptyText:  emptyText,
			Fixed:      true,
			Key:        "label",
			Label:      cputil.I18n(ctx, "label"),
			Options:    labelOptions,
			Type:       filter.PropConditionTypeSelect,
			HaveFilter: true,
		},
	)
	f.State.Conditions = propConditions
	urlParam, err := f.generateUrlQueryParams()
	if err != nil {
		return err
	}
	f.State.Base64UrlQueryParams = urlParam
	return nil
}

func (f *ComponentFilter) generateUrlQueryParams() (string, error) {
	fb, err := json.Marshal(f.State.Values)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(fb), nil
}

func (f *ComponentFilter) getPropIterationsOptions() (int64, []filter.PropConditionOption, error) {
	iterations, err := f.bdl.ListProjectIterations(apistructs.IterationPagingRequest{
		PageNo: 1, PageSize: 1000,
		ProjectID: f.projectID, State: apistructs.IterationStateUnfiled,
		WithoutIssueSummary: true,
	}, f.sdk.Identity.OrgID)
	if err != nil {
		return -1, nil, err
	}
	iterations = append(iterations, apistructs.Iteration{
		ID:    -1,
		Title: f.sdk.I18n("iterationUnassigned"),
	})
	var options []filter.PropConditionOption
	for _, iteration := range iterations {
		options = append(options, filter.PropConditionOption{
			Label: iteration.Title,
			Value: iteration.ID,
		})
	}
	// fixed iteration
	if f.fixedIterationID > 0 {
		found := false
		var fixedIteration apistructs.Iteration
		for i, itr := range iterations {
			if itr.ID == int64(f.fixedIterationID) {
				found = true
				fixedIteration = iterations[i]
				break
			}
		}
		if !found {
			return -1, nil, fmt.Errorf("fixedIteration: %d not belong to project", f.fixedIterationID)
		}
		options = []filter.PropConditionOption{{Label: fixedIteration.Title, Value: fixedIteration.ID}}
		return fixedIteration.ID, options, nil
	}
	return defaultIterationRetriever(iterations), options, nil
}

func (f *ComponentFilter) getProjectMemberOptions() ([]filter.PropConditionOption, error) {
	members, err := f.bdl.ListMembers(apistructs.MemberListRequest{
		ScopeType: apistructs.ProjectScope,
		ScopeID:   int64(f.projectID),
		PageNo:    1,
		PageSize:  500,
	})
	if err != nil {
		return nil, err
	}
	f.Members = members
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

func defaultIterationRetriever(iterations []apistructs.Iteration) int64 {
	for _, iteration := range iterations {
		if iteration.StartedAt != nil && !time.Now().Before(*iteration.StartedAt) &&
			iteration.FinishedAt != nil && !time.Now().After(*iteration.FinishedAt) {
			return iteration.ID
		}
	}
	return -1
}

func (f *ComponentFilter) getProjectLabelsOptions() ([]filter.PropConditionOption, error) {
	labels, err := f.bdl.Labels(string(apistructs.LabelTypeIssue), f.projectID, f.sdk.Identity.UserID)
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
		})
	}
	return options, nil
}
