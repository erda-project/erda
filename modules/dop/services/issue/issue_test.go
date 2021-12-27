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

// Package issue 封装 事件 相关操作
package issue

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/issuestream"
	"github.com/erda-project/erda/pkg/ucauth"
)

func TestCreateStream(t *testing.T) {
	streamFields := map[string][]interface{}{
		"title":            []interface{}{"a", "b"},
		"state":            []interface{}{1, 2},
		"plan_started_at":  []interface{}{"2021-12-01 00:00:00", "2021-12-02 00:00:00"},
		"plan_finished_at": []interface{}{"2021-12-01 00:00:00", "2021-12-02 00:00:00"},
		"owner":            []interface{}{"1", "2"},
		"stage":            []interface{}{"a", "b"},
		"priority":         []interface{}{apistructs.IssuePriorityLow, apistructs.IssuePriorityHigh},
		"complexity":       []interface{}{apistructs.IssueComplexityEasy, apistructs.IssueComplexityHard},
		"severity":         []interface{}{apistructs.IssueSeverityNormal, apistructs.IssueSeveritySerious},
		"content":          []interface{}{},
		"label":            []interface{}{},
		"assignee":         []interface{}{"1", "2"},
		"iteration_id":     []interface{}{1, 2},
	}
	db := &dao.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetIssueStateByID", func(client *dao.DBClient, ID int64) (*dao.IssueState, error) {
		return &dao.IssueState{Name: "a"}, nil
	})
	defer pm1.Unpatch()

	uc := &ucauth.UCClient{}
	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(uc), "FindUsers", func(c *ucauth.UCClient, ids []string) ([]ucauth.User, error) {
		return []ucauth.User{{Name: "a", Nick: "a"}, {Name: "b", Nick: "b"}}, nil
	})
	defer pm2.Unpatch()

	pm3 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetIssue", func(client *dao.DBClient, id int64) (dao.Issue, error) {
		return dao.Issue{Type: apistructs.IssueTypeTask}, nil
	})
	defer pm3.Unpatch()

	bdl := &bundle.Bundle{}
	pm4 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetProject", func(b *bundle.Bundle, id uint64) (*apistructs.ProjectDTO, error) {
		return &apistructs.ProjectDTO{OrgID: 1}, nil
	})
	defer pm4.Unpatch()

	pm5 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetIssuesStage", func(client *dao.DBClient, orgID int64, issueType apistructs.IssueType) ([]dao.IssueStage, error) {
		return []dao.IssueStage{{Name: "a"}}, nil
	})
	defer pm5.Unpatch()

	pm6 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetIteration", func(client *dao.DBClient, id uint64) (*dao.Iteration, error) {
		return &dao.Iteration{Title: "iteration"}, nil
	})
	defer pm6.Unpatch()

	stream := &issuestream.IssueStream{}
	pm7 := monkey.PatchInstanceMethod(reflect.TypeOf(stream), "Create", func(s *issuestream.IssueStream, req *apistructs.IssueStreamCreateRequest) (int64, error) {
		return 1, nil
	})
	defer pm7.Unpatch()

	svc := &Issue{db: db, uc: uc, bdl: bdl, stream: stream}
	err := svc.CreateStream(apistructs.IssueUpdateRequest{ID: 1, IdentityInfo: apistructs.IdentityInfo{UserID: "1"}}, streamFields)
	assert.NoError(t, err)
}

func Test_validPlanTime(t *testing.T) {
	type args struct {
		req   apistructs.IssueUpdateRequest
		issue *dao.Issue
	}
	timeBase := time.Date(2021, 9, 1, 0, 0, 0, 0, time.Now().Location())
	today := time.Date(2021, 9, 1, 0, 0, 0, 0, time.Now().Location())
	tomorrow := time.Date(2021, 9, 2, 0, 0, 0, 0, time.Now().Location())
	nilTime := time.Unix(0, 0)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				req: apistructs.IssueUpdateRequest{},
				issue: &dao.Issue{
					PlanStartedAt:  &timeBase,
					PlanFinishedAt: &timeBase,
				},
			},
		},
		{
			args: args{
				req: apistructs.IssueUpdateRequest{
					PlanStartedAt:  apistructs.IssueTime(tomorrow),
					PlanFinishedAt: apistructs.IssueTime(today),
				},
				issue: &dao.Issue{},
			},
			wantErr: true,
		},
		{
			args: args{
				req: apistructs.IssueUpdateRequest{
					PlanStartedAt: apistructs.IssueTime(nilTime),
				},
				issue: &dao.Issue{
					PlanStartedAt: &timeBase,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validPlanTime(tt.args.req, tt.args.issue); (err != nil) != tt.wantErr {
				t.Errorf("validPlanTime() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
