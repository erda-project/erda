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

package core

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	mttestplan "github.com/erda-project/erda/internal/apps/dop/services/testplan"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

func TestIssueService_GetTestPlanCaseRels(t *testing.T) {
	var db *dao.DBClient
	p1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "ListIssueTestCaseRelations",
		func(d *dao.DBClient, req apistructs.IssueTestCaseRelationsListRequest) ([]dao.IssueTestCaseRelation, error) {
			return []dao.IssueTestCaseRelation{
				{
					BaseModel: dbengine.BaseModel{
						ID: 1,
					},
				},
			}, nil
		},
	)
	defer p1.Unpatch()

	var mt *mttestplan.TestPlan
	p2 := monkey.PatchInstanceMethod(reflect.TypeOf(mt), "ListTestPlanCaseRels",
		func(d *mttestplan.TestPlan, req apistructs.TestPlanCaseRelListRequest) (rels []apistructs.TestPlanCaseRel, err error) {
			return []apistructs.TestPlanCaseRel{
				{
					ID: 1,
				},
			}, nil
		},
	)
	defer p2.Unpatch()

	svc := &IssueService{db: db, mttestPlan: mt}

	type args struct {
		issueID uint64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				issueID: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetTestPlanCaseRels(tt.args.issueID)
			if (err != nil) != tt.wantErr {
				t.Errorf("IssueService.GetTestPlanCaseRels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_getStage(t *testing.T) {
	type args struct {
		req *pb.IssueCreateRequest
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				req: &pb.IssueCreateRequest{
					Type:     pb.IssueTypeEnum_TASK,
					TaskType: "a",
				},
			},
			want: "a",
		},
		{
			args: args{
				req: &pb.IssueCreateRequest{
					Type:     pb.IssueTypeEnum_BUG,
					BugStage: "b",
				},
			},
			want: "b",
		},
		{
			args: args{
				req: &pb.IssueCreateRequest{
					Type: pb.IssueTypeEnum_REQUIREMENT,
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStage(tt.args.req); got != tt.want {
				t.Errorf("getStage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIssueService_DBClient(t *testing.T) {
	db := &dao.DBClient{}
	i := IssueService{db: db}
	assert.Equal(t, db, i.DBClient())
}

func TestIssueService_validateIssueCreateRequest_TimeOrder(t *testing.T) {
	svc := &IssueService{}
	start := time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)
	finish := time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)

	err := svc.validateIssueCreateRequest(&pb.IssueCreateRequest{
		ProjectID:   2,
		IterationID: -1,
	}, &start, &finish)
	if err == nil {
		t.Fatal("expected validation error when planStartedAt > planFinishedAt")
	}
}

func TestIssueService_validateIssueCreateRequest_IterationNotFound(t *testing.T) {
	var db *dao.DBClient
	patch := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetIteration",
		func(_ *dao.DBClient, _ uint64) (*dao.Iteration, error) {
			return nil, gorm.ErrRecordNotFound
		},
	)
	defer patch.Unpatch()

	svc := &IssueService{db: db}
	err := svc.validateIssueCreateRequest(&pb.IssueCreateRequest{
		ProjectID:   2,
		IterationID: 999999,
	}, nil, nil)
	if err == nil {
		t.Fatal("expected validation error for non-existing iteration")
	}
}

func TestIssueService_validateIssueCreateRequest_IterationProjectMismatch(t *testing.T) {
	var db *dao.DBClient
	patch := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetIteration",
		func(_ *dao.DBClient, _ uint64) (*dao.Iteration, error) {
			return &dao.Iteration{ProjectID: 3}, nil
		},
	)
	defer patch.Unpatch()

	svc := &IssueService{db: db}
	err := svc.validateIssueCreateRequest(&pb.IssueCreateRequest{
		ProjectID:   2,
		IterationID: 3,
	}, nil, nil)
	if err == nil {
		t.Fatal("expected validation error for mismatched project iteration")
	}
}

func TestIssueService_validateIssueCreateRequest_IterationDBError(t *testing.T) {
	var db *dao.DBClient
	patch := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetIteration",
		func(_ *dao.DBClient, _ uint64) (*dao.Iteration, error) {
			return nil, errors.New("db unavailable")
		},
	)
	defer patch.Unpatch()

	svc := &IssueService{db: db}
	err := svc.validateIssueCreateRequest(&pb.IssueCreateRequest{
		ProjectID:   2,
		IterationID: 3,
	}, nil, nil)
	if err == nil {
		t.Fatal("expected internal error when iteration query fails")
	}
}
