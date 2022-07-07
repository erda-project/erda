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
	"io"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	issuepb "github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	syncpb "github.com/erda-project/erda-proto-go/dop/issue/sync/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
)

type issueMock struct{}

func (i issueMock) Paging(req issuepb.PagingIssueRequest) ([]*issuepb.Issue, uint64, error) {
	panic("implement me")
}

func (i issueMock) GetIssue(id int64, identityInfo *commonpb.IdentityInfo) (*issuepb.Issue, error) {
	panic("implement me")
}

func (i issueMock) AfterIssueUpdate(u *query.IssueUpdated) error {
	panic("implement me")
}

func (i issueMock) UpdateIssue(req *issuepb.UpdateIssueRequest) error {
	panic("implement me")
}

func (i issueMock) UpdateLabels(id, projectID uint64, labelNames []string) (err error) {
	panic("implement me")
}

func (i issueMock) SyncLabels(value *syncpb.Value, issueIDs []uint64) error {
	panic("implement me")
}

func (i issueMock) BatchUpdateIssue(req *issuepb.BatchUpdateIssueRequest) error {
	panic("implement me")
}

func (i issueMock) GetProperties(req *issuepb.GetIssuePropertyRequest) ([]*issuepb.IssuePropertyIndex, error) {
	panic("implement me")
}

func (i issueMock) CreatePropertyRelation(req *issuepb.CreateIssuePropertyInstanceRequest) error {
	panic("implement me")
}

func (i issueMock) GetIssueStage(req *issuepb.IssueStageRequest) ([]*issuepb.IssueStage, error) {
	panic("implement me")
}

func (i issueMock) GetIssueRelationsByIssueIDs(issueID uint64, relationType []string) ([]uint64, []uint64, error) {
	panic("implement me")
}

func (i issueMock) GetIssuesByIssueIDs(issueIDs []uint64) ([]*issuepb.Issue, error) {
	panic("implement me")
}

func (i issueMock) GetBatchProperties(orgID int64, issuesType []string) ([]*issuepb.IssuePropertyIndex, error) {
	panic("implement me")
}

func (i issueMock) ExportExcel(issues []*issuepb.Issue, properties []*issuepb.IssuePropertyIndex, projectID uint64, isDownload bool, orgID int64, locale string) (io.Reader, string, error) {
	panic("implement me")
}

func (i issueMock) SyncIssueChildrenIteration(issue *issuepb.Issue, iterationID int64) error {
	panic("implement me")
}

func (i issueMock) AfterIssueAppRelationCreate(issueIDs []int64) error {
	panic("implement me")
}

func (i issueMock) GetIssueLabelsByProjectID(projectID uint64) ([]dao.IssueLabel, error) {
	panic("implement me")
}

func (i issueMock) GetIssuesStatesByProjectID(projectID uint64, issueType string) ([]dao.IssueState, error) {
	panic("implement me")
}

func (i issueMock) GetAllIssuesByProject(req issuepb.IssueListRequest) ([]dao.IssueItem, error) {
	panic("implement me")
}

func (i issueMock) GetIssueChildren(id uint64, req issuepb.PagingIssueRequest) ([]dao.IssueItem, uint64, error) {
	panic("implement me")
}

func (i issueMock) GetIssueItem(id uint64) (*dao.IssueItem, error) {
	panic("implement me")
}

func (i issueMock) GetIssueParents(issueID uint64, relationType []string) ([]dao.IssueItem, error) {
	panic("implement me")
}

func (i issueMock) ListStatesTransByProjectID(projectID uint64) ([]dao.IssueStateTransition, error) {
	panic("implement me")
}

func (i issueMock) GetIssueStateIDsByTypes(req *apistructs.IssueStatesRequest) ([]int64, error) {
	panic("implement me")
}

func (i issueMock) GetIssueStatesMap(req *issuepb.GetIssueStatesRequest) (map[string][]issuepb.IssueStatus, error) {
	panic("implement me")
}

func (i issueMock) GetIssueStateIDs(req *issuepb.GetIssueStatesRequest) ([]int64, error) {
	panic("implement me")
}

func (i issueMock) GetIssueStatesBelong(req *issuepb.GetIssueStateRelationRequest) ([]apistructs.IssueStateState, error) {
	panic("implement me")
}

func (i issueMock) AfterIssueInclusionRelationChange(id uint64) error {
	panic("implement me")
}
