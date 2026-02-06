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

//go:generate mockgen -destination=./mock_usersvc_test.go -package core github.com/erda-project/erda-proto-go/core/user/pb UserServiceServer

package core

import (
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"
	"github.com/jinzhu/gorm"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/stream/common"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/discover"
)

func Test_provider_CreateStream(t *testing.T) {
	startedAt := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
	finishedAt := time.Date(2021, 12, 2, 0, 0, 0, 0, time.UTC)
	streamFields := map[string][]interface{}{
		"title":            {"a", "b"},
		"state":            {int64(1), int64(2)},
		"plan_started_at":  {&startedAt, &finishedAt},
		"plan_finished_at": {&startedAt, &finishedAt},
		"owner":            {"1", "2"},
		"priority":         {string(apistructs.IssuePriorityLow), string(apistructs.IssuePriorityHigh)},
		"complexity":       {string(apistructs.IssueComplexityEasy), string(apistructs.IssueComplexityHard)},
		"severity":         {string(apistructs.IssueSeverityNormal), string(apistructs.IssueSeveritySerious)},
		"content":          {},
		"label":            {},
		"assignee":         {"1", "2"},
		"iteration_id":     {int64(1), int64(2)},
	}
	db, mock, cleanup := setupMockDBClient(t)
	defer cleanup()
	mock.MatchExpectationsInOrder(false)

	patches := gomonkey.NewPatches()
	patches.ApplyFunc(discover.GetEndpoint, func(string) (string, error) { return "localhost:9093", nil })
	defer patches.Reset()

	mock.ExpectQuery("SELECT .* FROM `dice_issue_state`.*").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "a"))
	mock.ExpectQuery("SELECT .* FROM `dice_issue_state`.*").
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(2, "b"))
	mock.ExpectQuery("SELECT .* FROM `dice_iterations`.*").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(1, "1"))
	mock.ExpectQuery("SELECT .* FROM `dice_iterations`.*").
		WithArgs(uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(2, "2"))

	for i := 0; i < len(streamFields); i++ {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `dice_issue_streams` .*").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
	}
	mock.ExpectClose()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userSvc := NewMockUserServiceServer(ctrl)
	userSvc.EXPECT().FindUsers(gomock.Any(), gomock.Any()).AnyTimes().Return(&userpb.FindUsersResponse{Data: []*commonpb.UserInfo{{Name: "a", Nick: "a"}, {Name: "b", Nick: "b"}}}, nil)

	bdl := &bundle.Bundle{}
	patches.ApplyMethod(reflect.TypeOf(bdl), "GetProject", func(b *bundle.Bundle, id uint64) (*apistructs.ProjectDTO, error) {
		return &apistructs.ProjectDTO{OrgID: 1}, nil
	})

	p := &provider{db: db, bdl: bdl, UserSvc: userSvc, I18n: &mockTranslator{}}

	done := make(chan struct{}, 1)
	patches.ApplyMethod(reflect.TypeOf(p), "CreateIssueEvent", func(_ *provider, _ *common.IssueStreamCreateRequest) error {
		done <- struct{}{}
		return nil
	})
	defer patches.Reset()

	err := p.CreateStream(&pb.UpdateIssueRequest{Id: 1, IdentityInfo: &commonpb.IdentityInfo{UserID: "1"}}, streamFields)
	assert.NoError(t, err)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("CreateIssueEvent was not called")
	}
}

func Test_provider_HandleIssueStreamChangeIteration(t *testing.T) {
	// Mock iteration lookups with sqlmock.
	db, mock, cleanup := setupMockDBClient(t)
	defer cleanup()
	mock.ExpectQuery("SELECT .* FROM `dice_iterations`.*").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(1, "1"))
	mock.ExpectQuery("SELECT .* FROM `dice_iterations`.*").
		WithArgs(uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(2, "2"))
	mock.ExpectQuery("SELECT .* FROM `dice_iterations`.*").
		WithArgs(uint64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(3, "3"))
	mock.ExpectQuery("SELECT .* FROM `dice_iterations`.*").
		WithArgs(uint64(4)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(4, "4"))
	mock.ExpectClose()

	svc := &provider{db: db, I18n: &mockTranslator{}}

	// From unassigned to concrete iteration.
	streamType, params, err := svc.HandleIssueStreamChangeIteration(nil, apistructs.UnassignedIterationID, 1)
	assert.NoError(t, err)
	assert.Equal(t, common.ISTChangeIterationFromUnassigned, streamType)
	assert.Equal(t, "1", params.NewIteration)

	// From concrete iteration to unassigned.
	streamType, params, err = svc.HandleIssueStreamChangeIteration(nil, 2, apistructs.UnassignedIterationID)
	assert.NoError(t, err)
	assert.Equal(t, common.ISTChangeIterationToUnassigned, streamType)
	assert.Equal(t, "2", params.CurrentIteration)

	// From concrete to concrete iteration.
	streamType, params, err = svc.HandleIssueStreamChangeIteration(nil, 3, 4)
	assert.NoError(t, err)
	assert.Equal(t, common.ISTChangeIteration, streamType)
	assert.Equal(t, "3", params.CurrentIteration)
	assert.Equal(t, "4", params.NewIteration)
}

func Test_provider_CreateIssueStreamBySystem(t *testing.T) {
	startedAt := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
	finishedAt := time.Date(2021, 12, 2, 0, 0, 0, 0, time.UTC)

	db, mock, cleanup := setupMockDBClient(t)
	defer cleanup()
	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery("SELECT .* FROM `dice_issue_state`.*").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "open"))
	mock.ExpectQuery("SELECT .* FROM `dice_issue_state`.*").
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(2, "closed"))
	mock.ExpectQuery("SELECT .* FROM `dice_iterations`.*").
		WithArgs(uint64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(3, "3"))
	mock.ExpectQuery("SELECT .* FROM `dice_iterations`.*").
		WithArgs(uint64(4)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(4, "4"))
	mock.ExpectExec("INSERT INTO `dice_issue_streams` .*").
		WillReturnResult(sqlmock.NewResult(1, 5))
	mock.ExpectClose()

	p := &provider{db: db, I18n: &mockTranslator{}}
	streamFields := map[string][]interface{}{
		"state":            {int64(1), int64(2), "system"},
		"plan_started_at":  {&startedAt, &finishedAt, "system"},
		"plan_finished_at": {&startedAt, &finishedAt, "system"},
		"label":            {"", "", "system"},
		"iteration_id":     {int64(3), int64(4), "system"},
	}
	err := p.CreateIssueStreamBySystem(1, streamFields)
	assert.NoError(t, err)
}

func setupMockDBClient(t *testing.T) (*dao.DBClient, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	assert.NoError(t, err)

	gdb, err := gorm.Open("mysql", db)
	assert.NoError(t, err)
	gdb.LogMode(false)

	client := &dao.DBClient{DBEngine: &dbengine.DBEngine{DB: gdb}}
	cleanup := func() {
		assert.NoError(t, gdb.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
	}
	return client, mock, cleanup
}
