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

package flow

import (
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	issuepb "github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	syncpb "github.com/erda-project/erda-proto-go/dop/issue/sync/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
)

type IssueMock struct{}

func (i IssueMock) Paging(req issuepb.PagingIssueRequest) ([]*issuepb.Issue, uint64, error) {
	panic("implement me")
}

func (i IssueMock) GetIssue(id int64, identityInfo *commonpb.IdentityInfo) (*issuepb.Issue, error) {
	panic("implement me")
}

func (i IssueMock) AfterIssueUpdate(u *query.IssueUpdated) error {
	panic("implement me")
}

func (i IssueMock) UpdateIssue(req *issuepb.UpdateIssueRequest) error {
	panic("implement me")
}

func (i IssueMock) UpdateLabels(id, projectID uint64, labelNames []string) (err error) {
	panic("implement me")
}

func (i IssueMock) SyncLabels(value *syncpb.Value, issueIDs []uint64) error {
	panic("implement me")
}

func (i IssueMock) BatchUpdateIssue(req *issuepb.BatchUpdateIssueRequest) error {
	panic("implement me")
}

func (i IssueMock) GetProperties(req *issuepb.GetIssuePropertyRequest) ([]*issuepb.IssuePropertyIndex, error) {
	panic("implement me")
}

func (i IssueMock) CreatePropertyRelation(req *issuepb.CreateIssuePropertyInstanceRequest) error {
	panic("implement me")
}

func (i IssueMock) GetIssueStage(req *issuepb.IssueStageRequest) ([]*issuepb.IssueStage, error) {
	panic("implement me")
}

func (i IssueMock) GetIssueRelationsByIssueIDs(issueID uint64, relationType []string) ([]uint64, []uint64, error) {
	panic("implement me")
}

func (i IssueMock) GetIssuesByIssueIDs(issueIDs []uint64) ([]*issuepb.Issue, error) {
	panic("implement me")
}

func (i IssueMock) GetBatchProperties(orgID int64, issuesType []string) ([]*issuepb.IssuePropertyIndex, error) {
	panic("implement me")
}

func (i IssueMock) SyncIssueChildrenIteration(issue *issuepb.Issue, iterationID int64) error {
	panic("implement me")
}

func (i IssueMock) AfterIssueAppRelationCreate(issueIDs []int64) error {
	panic("implement me")
}

func (i IssueMock) GetIssueLabelsByProjectID(projectID uint64) ([]dao.IssueLabel, error) {
	panic("implement me")
}

func (i IssueMock) GetIssuesStatesByProjectID(projectID uint64, issueType string) ([]dao.IssueState, error) {
	panic("implement me")
}

func (i IssueMock) GetAllIssuesByProject(req issuepb.IssueListRequest) ([]dao.IssueItem, error) {
	panic("implement me")
}

func (i IssueMock) GetIssueChildren(id uint64, req issuepb.PagingIssueRequest) ([]dao.IssueItem, uint64, error) {
	panic("implement me")
}

func (i IssueMock) GetIssueItem(id uint64) (*dao.IssueItem, error) {
	panic("implement me")
}

func (i IssueMock) GetIssueParents(issueID uint64, relationType []string) ([]dao.IssueItem, error) {
	panic("implement me")
}

func (i IssueMock) ListStatesTransByProjectID(projectID uint64) ([]dao.IssueStateTransition, error) {
	panic("implement me")
}

func (i IssueMock) GetIssueStateIDsByTypes(req *apistructs.IssueStatesRequest) ([]int64, error) {
	panic("implement me")
}

func (i IssueMock) GetIssueStatesMap(req *issuepb.GetIssueStatesRequest) (map[string][]issuepb.IssueStatus, error) {
	panic("implement me")
}

func (i IssueMock) GetIssueStateIDs(req *issuepb.GetIssueStatesRequest) ([]int64, error) {
	panic("implement me")
}

func (i IssueMock) GetIssueStatesBelong(req *issuepb.GetIssueStateRelationRequest) ([]apistructs.IssueStateState, error) {
	panic("implement me")
}

func (i IssueMock) AfterIssueInclusionRelationChange(id uint64) error {
	panic("implement me")
}

func (i IssueMock) BatchGetProperties(orgID int64, issuesType []string) ([]*issuepb.IssuePropertyIndex, error) {
	panic("implement me")
}

func (i IssueMock) GetIssuePropertyInstance(req *issuepb.GetIssuePropertyInstanceRequest) (*issuepb.IssueAndPropertyAndValue, error) {
	panic("implement me")
}

func (i IssueMock) BatchGetIssuePropertyInstances(orgID int64, projectID uint64, issueType string, issueIDs []uint64) (map[uint64]*issuepb.IssueAndPropertyAndValue, error) {
	panic("implement me")
}

func (i IssueMock) BatchGetIssue(id []int64, identityInfo *commonpb.IdentityInfo) ([]*issuepb.Issue, error) {
	panic("implement me")
}
