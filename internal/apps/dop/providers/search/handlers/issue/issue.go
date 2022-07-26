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

package issue

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	issuepb "github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda-proto-go/dop/search/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/apps/dop/providers/search/handlers"
)

type issueSearch struct {
	handlers.BaseSearch

	bdl   *bundle.Bundle
	query query.Interface
}

var (
	projectLimit = 50
)

type Option func(*issueSearch)

func WithQuery(query query.Interface) Option {
	return func(i *issueSearch) {
		i.query = query
	}
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(i *issueSearch) {
		i.bdl = bdl
	}
}

func WithNexts(nexts ...handlers.Handler) Option {
	return func(i *issueSearch) {
		i.Nexts = nexts
	}
}

func (i *issueSearch) BeginSearch(ctx context.Context, req *pb.SearchRequest) {
	defer i.DoNexts(ctx, req)
	// get joined projects
	projects, err := i.bdl.ListMyProject(req.IdentityInfo.UserID, apistructs.ProjectListRequest{
		OrgID:    req.OrgID,
		PageSize: projectLimit,
		Joined:   true,
		PageNo:   1,
	})
	if err != nil {
		i.AppendError(err)
		return
	}
	if projects == nil {
		return
	}
	projectIDs := make([]uint64, 0)
	for _, project := range projects.List {
		projectIDs = append(projectIDs, project.ID)
	}
	if len(projectIDs) == 0 {
		return
	}
	issues, _, err := i.query.Paging(issuepb.PagingIssueRequest{
		Title:      req.Query,
		OrgID:      int64(req.OrgID),
		PageSize:   handlers.PageSize,
		External:   true,
		ProjectIDs: projectIDs,
	})
	if err != nil {
		i.AppendError(err)
		return
	}
	res := &pb.SearchResultContent{
		Type: handlers.SearchTypeIssue.String(),
	}
	org, err := i.bdl.GetOrg(req.OrgID)
	if err != nil {
		i.AppendError(err)
		return
	}
	for _, issue := range issues {
		data, err := i.convertIssueToValue(issue)
		if err != nil {
			i.AppendError(err)
			continue
		}
		res.Items = append(res.Items, &pb.SearchResultItem{
			Item: data,
			Link: i.genIssueLink(org.Name, issue),
		})
	}
	i.AppendContent(res)
}

// convertIssueToValue converts issue to search value
// and filters out the fields that are not needed like content.
func (i *issueSearch) convertIssueToValue(issue *issuepb.Issue) (*structpb.Value, error) {
	issue.Content = ""
	issueBytes, err := issue.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var issueMap map[string]interface{}
	if err := json.Unmarshal(issueBytes, &issueMap); err != nil {
		return nil, err
	}
	return structpb.NewValue(issueMap)
}

func (i *issueSearch) genIssueLink(orgName string, issue *issuepb.Issue) string {
	return fmt.Sprintf("/%s/dop/projects/%d/issues/all?id=%d&type=%s", orgName, issue.ProjectID, issue.Id, issue.Type)
}

func NewIssueHandler(opts ...Option) *issueSearch {
	i := &issueSearch{}
	for _, op := range opts {
		op(i)
	}
	return i
}
