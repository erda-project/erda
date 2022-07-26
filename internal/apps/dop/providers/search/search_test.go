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

package search

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda-proto-go/common/pb"
	pb0 "github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	pb1 "github.com/erda-project/erda-proto-go/dop/issue/sync/pb"
	pb2 "github.com/erda-project/erda-proto-go/dop/search/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/internal/apps/dop/providers/search/handlers"
	"github.com/erda-project/erda/pkg/common/apis"
)

func TestSearch(t *testing.T) {
	bdl := bundle.New()
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ListMyProject", func(_ *bundle.Bundle, userID string, req apistructs.ProjectListRequest) (*apistructs.PagingProjectDTO, error) {
		return &apistructs.PagingProjectDTO{
			Total: 2,
			List: []apistructs.ProjectDTO{
				{
					ID:   1,
					Name: "project1",
				},
				{
					ID:   2,
					Name: "project2",
				},
			},
		}, nil
	})
	defer pm1.Unpatch()

	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetOrg", func(_ *bundle.Bundle, idOrName interface{}) (*apistructs.OrgDTO, error) {
		return &apistructs.OrgDTO{
			ID:   1,
			Name: "org1",
		}, nil
	})
	defer pm2.Unpatch()

	pm3 := monkey.Patch(apis.GetIdentityInfo, func(_ context.Context) *pb.IdentityInfo {
		return &pb.IdentityInfo{UserID: "1"}
	})
	defer pm3.Unpatch()

	pm4 := monkey.Patch(apis.GetIntOrgID, func(_ context.Context) (int64, error) {
		return 1, nil
	})
	defer pm4.Unpatch()

	pm5 := monkey.Patch(apis.GetOrgID, func(_ context.Context) string {
		return "1"
	})
	defer pm5.Unpatch()

	p := &provider{
		Query: &MockIssueQuery{},
		bdl:   bdl,
		Cfg:   &config{},
	}
	service := &ServiceImpl{
		query: p.Query,
		bdl:   p.bdl,
	}
	res, err := service.Search(context.Background(), &pb2.SearchRequest{
		Query: "issue",
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res.Data))
	assert.Equal(t, handlers.SearchTypeIssue.String(), res.Data[0].Type)
	assert.Equal(t, 2, len(res.Data[0].Items))
}

// MockIssueQuery is a mock of Interface interface.
type MockIssueQuery struct{}

func (m *MockIssueQuery) AfterIssueAppRelationCreate(arg0 []int64) error {
	panic("implement me")
}

func (m *MockIssueQuery) AfterIssueInclusionRelationChange(arg0 uint64) error {
	panic("implement me")
}

func (m *MockIssueQuery) AfterIssueUpdate(arg0 *query.IssueUpdated) error {
	panic("implement me")
}

func (m *MockIssueQuery) BatchUpdateIssue(arg0 *pb0.BatchUpdateIssueRequest) error {
	panic("implement me")
}

func (m *MockIssueQuery) CreatePropertyRelation(arg0 *pb0.CreateIssuePropertyInstanceRequest) error {
	panic("implement me")
}

func (m *MockIssueQuery) ExportExcel(arg0 []*pb0.Issue, arg1 []*pb0.IssuePropertyIndex, arg2 uint64, arg3 bool, arg4 int64, arg5 string) (io.Reader, string, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetAllIssuesByProject(arg0 pb0.IssueListRequest) ([]dao.IssueItem, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetBatchProperties(arg0 int64, arg1 []string) ([]*pb0.IssuePropertyIndex, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetIssue(arg0 int64, arg1 *pb.IdentityInfo) (*pb0.Issue, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetIssueChildren(arg0 uint64, arg1 pb0.PagingIssueRequest) ([]dao.IssueItem, uint64, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetIssueItem(arg0 uint64) (*dao.IssueItem, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetIssueLabelsByProjectID(arg0 uint64) ([]dao.IssueLabel, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetIssueParents(arg0 uint64, arg1 []string) ([]dao.IssueItem, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetIssueRelationsByIssueIDs(arg0 uint64, arg1 []string) ([]uint64, []uint64, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetIssueStage(arg0 *pb0.IssueStageRequest) ([]*pb0.IssueStage, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetIssueStateIDs(arg0 *pb0.GetIssueStatesRequest) ([]int64, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetIssueStateIDsByTypes(arg0 *apistructs.IssueStatesRequest) ([]int64, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetIssueStatesBelong(arg0 *pb0.GetIssueStateRelationRequest) ([]apistructs.IssueStateState, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetIssueStatesMap(arg0 *pb0.GetIssueStatesRequest) (map[string][]pb0.IssueStatus, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetIssuesByIssueIDs(arg0 []uint64) ([]*pb0.Issue, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetIssuesStatesByProjectID(arg0 uint64, arg1 string) ([]dao.IssueState, error) {
	panic("implement me")
}

func (m *MockIssueQuery) GetProperties(arg0 *pb0.GetIssuePropertyRequest) ([]*pb0.IssuePropertyIndex, error) {
	panic("implement me")
}

func (m *MockIssueQuery) ListStatesTransByProjectID(arg0 uint64) ([]dao.IssueStateTransition, error) {
	panic("implement me")
}

func (m *MockIssueQuery) Paging(arg0 pb0.PagingIssueRequest) ([]*pb0.Issue, uint64, error) {
	res := make([]*pb0.Issue, 0)
	var total uint64
	for _, projectID := range arg0.ProjectIDs {
		res = append(res, &pb0.Issue{ProjectID: projectID, Title: fmt.Sprintf("%d", projectID)})
		total++
	}
	return res, total, nil
}

func (m *MockIssueQuery) SyncIssueChildrenIteration(arg0 *pb0.Issue, arg1 int64) error {
	panic("implement me")
}

func (m *MockIssueQuery) SyncLabels(arg0 *pb1.Value, arg1 []uint64) error {
	panic("implement me")
}

func (m *MockIssueQuery) UpdateIssue(arg0 *pb0.UpdateIssueRequest) error {
	panic("implement me")
}

func (m *MockIssueQuery) UpdateLabels(arg0, arg1 uint64, arg2 []string) error {
	panic("implement me")
}
