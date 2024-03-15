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

package query

import (
	"reflect"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	syncpb "github.com/erda-project/erda-proto-go/dop/issue/sync/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	stream "github.com/erda-project/erda/internal/apps/dop/providers/issue/stream/core"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type config struct {
}

type provider struct {
	Cfg      *config
	Log      logs.Logger
	DB       *gorm.DB `autowired:"mysql-client"`
	bdl      *bundle.Bundle
	db       *dao.DBClient
	Stream   stream.Interface
	Identity userpb.UserServiceServer
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithErdaServer())
	p.db = &dao.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: p.DB,
		},
	}
	return nil
}

// Interface .
// How to generate mock code:
//
//	mockgen -destination pkg/mock/issue_query_mock.go -package mock -source internal/apps/dop/providers/issue/core/query/provider.go -mock_names Interface=MockIssueQuery
type Interface interface {
	Paging(req pb.PagingIssueRequest) ([]*pb.Issue, uint64, error)
	GetIssue(id int64, identityInfo *commonpb.IdentityInfo) (*pb.Issue, error)
	AfterIssueUpdate(u *IssueUpdated) error
	UpdateIssue(req *pb.UpdateIssueRequest) error
	UpdateLabels(id, projectID uint64, labelNames []string) (err error)
	SyncLabels(value *syncpb.Value, issueIDs []uint64) error
	BatchUpdateIssue(req *pb.BatchUpdateIssueRequest) error
	GetProperties(req *pb.GetIssuePropertyRequest) ([]*pb.IssuePropertyIndex, error)
	BatchGetProperties(orgID int64, issuesType []string) ([]*pb.IssuePropertyIndex, error)
	CreatePropertyRelation(req *pb.CreateIssuePropertyInstanceRequest) error
	GetIssuePropertyInstance(req *pb.GetIssuePropertyInstanceRequest) (*pb.IssueAndPropertyAndValue, error)
	BatchGetIssuePropertyInstances(orgID int64, projectID uint64, issueType string, issueIDs []uint64) (map[uint64]*pb.IssueAndPropertyAndValue, error)
	GetIssueStage(req *pb.IssueStageRequest) ([]*pb.IssueStage, error)
	GetIssueRelationsByIssueIDs(issueID uint64, relationType []string) ([]uint64, []uint64, error)
	GetIssuesByIssueIDs(issueIDs []uint64) ([]*pb.Issue, error)
	SyncIssueChildrenIteration(issue *pb.Issue, iterationID int64) error
	AfterIssueAppRelationCreate(issueIDs []int64) error
	GetIssueLabelsByProjectID(projectID uint64) ([]dao.IssueLabel, error)
	GetIssuesStatesByProjectID(projectID uint64, issueType string) ([]dao.IssueState, error)
	GetAllIssuesByProject(req pb.IssueListRequest) ([]dao.IssueItem, error)
	GetIssueChildren(id uint64, req pb.PagingIssueRequest) ([]dao.IssueItem, uint64, error)
	GetIssueItem(id uint64) (*dao.IssueItem, error)
	GetIssueParents(issueID uint64, relationType []string) ([]dao.IssueItem, error)
	ListStatesTransByProjectID(projectID uint64) ([]dao.IssueStateTransition, error)
	GetIssueStateIDsByTypes(req *apistructs.IssueStatesRequest) ([]int64, error)
	GetIssueStatesMap(req *pb.GetIssueStatesRequest) (map[string][]pb.IssueStatus, error)
	GetIssueStateIDs(req *pb.GetIssueStatesRequest) ([]int64, error)
	GetIssueStatesBelong(req *pb.GetIssueStateRelationRequest) ([]apistructs.IssueStateState, error)
	AfterIssueInclusionRelationChange(id uint64) error
}

func init() {
	servicehub.Register("erda.dop.issue.core.query", &servicehub.Spec{
		Services:   []string{"erda.dop.issue.core.query"},
		Types:      []reflect.Type{reflect.TypeOf((*Interface)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
