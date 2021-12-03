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

package gantt

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/issue"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/pkg/strutil"
)

func init() {
	base.InitProviderWithCreator("issue-gantt", "gantt",
		func() servicehub.Provider { return &ComponentGantt{} })
}

func (f *ComponentGantt) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	f.sdk = cputil.SDK(ctx)
	f.bdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	f.issueSvc = ctx.Value(types.IssueService).(*issue.Issue)
	f.users = make([]string, 0)
	projectID, err := strconv.ParseUint(cputil.GetInParamByKey(ctx, "projectId").(string), 10, 64)
	if err != nil {
		return err
	}
	f.projectID = projectID
	var op OperationData
	dataBody, err := json.Marshal(event.OperationData)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(dataBody, &op); err != nil {
		return err
	}

	req := apistructs.IssuePagingRequest{
		IssueListRequest: apistructs.IssueListRequest{
			ProjectID:    f.projectID,
			Type:         []apistructs.IssueType{apistructs.IssueTypeRequirement, apistructs.IssueTypeTask},
			IterationIDs: f.State.Values.IterationIDs,
			Label:        f.State.Values.LabelIDs,
			Assignees:    f.State.Values.AssigneeIDs,
		},
		PageNo:   1,
		PageSize: 500,
	}

	expand := make(map[uint64][]Item)
	switch event.Operation {
	case cptype.InitializeOperation, cptype.RenderingOperation:
		f.Operations = map[apistructs.OperationKey]Operation{
			apistructs.Update: {
				Key:      apistructs.Update.String(),
				Reload:   true,
				FillMeta: "nodes",
				Async:    true,
			},
			apistructs.ExpandNode: {
				Key:      apistructs.ExpandNode.String(),
				Reload:   true,
				FillMeta: "keys",
			},
		}

		issues, _, err := f.issueSvc.GetIssueChildren(0, req)
		if err != nil {
			return err
		}
		expand[0] = f.convertIssueItem(issues)
	case cptype.OperationKey(apistructs.ExpandNode):
		for _, key := range op.Meta.Keys {
			issues, _, err := f.issueSvc.GetIssueChildren(key, req)
			if err != nil {
				return err
			}
			expand[key] = f.convertIssueItem(issues)
		}
	case cptype.OperationKey(apistructs.Update):
		id := op.Meta.Nodes.Key
		if err := f.issueSvc.UpdateIssue(apistructs.IssueUpdateRequest{
			ID:             id,
			PlanStartedAt:  timeFromMilli(op.Meta.Nodes.Start),
			PlanFinishedAt: timeFromMilli(op.Meta.Nodes.End),
		}); err != nil {
			return err
		}

		parents, err := f.issueSvc.GetIssueParents(id, []string{apistructs.IssueRelationInclusion})
		if err != nil {
			return err
		}
		issue, err := f.issueSvc.GetIssueItem(id)
		if err != nil {
			return err
		}
		f.Data.UpdateList = append(f.convertIssueItem(parents),
			Item{
				Title:  issue.Title,
				Key:    uint64(issue.ID),
				IsLeaf: issue.Type == apistructs.IssueTypeTask,
				Start:  issue.PlanStartedAt,
				End:    issue.PlanFinishedAt,
				Extra: Extra{
					Type: issue.Type.String(),
					User: issue.Assignee,
					Status: Status{
						Text:   issue.Name,
						Status: apistructs.IssueStateBelong(issue.Belong).GetFrontEndStatus(),
					},
				},
			},
		)
	}

	f.Data.ExpandList = expand
	(*gs)[protocol.GlobalInnerKeyUserIDs.String()] = strutil.DedupSlice(f.users)
	return nil
}

func (f *ComponentGantt) convertIssueItem(issues []dao.IssueItem) []Item {
	res := make([]Item, 0, len(issues))
	for _, issue := range issues {
		item := Item{
			Title:  issue.Title,
			Key:    issue.ID,
			IsLeaf: issue.Type == apistructs.IssueTypeTask,
			Extra: Extra{
				Type: issue.Type.String(),
				User: issue.Assignee,
				Status: Status{
					Text:   issue.Name,
					Status: apistructs.IssueStateBelong(issue.Belong).GetFrontEndStatus(),
				},
			},
		}
		if issue.PlanStartedAt != nil {
			item.Start = issue.PlanStartedAt
		}
		if issue.PlanFinishedAt != nil {
			item.End = issue.PlanFinishedAt
		}
		res = append(res, item)
		f.users = append(f.users, issue.Assignee)
	}
	return res
}

func timeFromMilli(millis int64) *time.Time {
	t := time.Unix(0, millis*int64(time.Millisecond))
	return &t
}

func milliFromTime(t *time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
