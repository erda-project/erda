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
	"context"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/alecthomas/assert"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/stream/common"
	"github.com/erda-project/erda/internal/pkg/mock"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type EventOrgMock struct {
	mock.OrgMock
}

func (m EventOrgMock) GetOrg(ctx context.Context, request *orgpb.GetOrgRequest) (*orgpb.GetOrgResponse, error) {
	return &orgpb.GetOrgResponse{Data: &orgpb.Org{ID: 1}}, nil
}

func Test_filterReceiversByOperatorID(t *testing.T) {
	type args struct {
		receivers  []string
		operatorID string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			args: args{
				receivers:  []string{"a", "b"},
				operatorID: "b",
			},
			want: []string{"a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterReceiversByOperatorID(tt.args.receivers, tt.args.operatorID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterReceiversByOperatorID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_groupEventContent(t *testing.T) {
	content, err := groupEventContent([]string{common.ISTChangeContent}, common.ISTParam{}, &mockTranslator{}, "en")
	assert.NoError(t, err)
	assert.Equal(t, "content changed", content)
}

func Test_provider_CreateIssueEvent(t *testing.T) {
	db, mock, cleanup := setupMockDBClient(t)
	defer cleanup()
	mock.ExpectQuery("SELECT .* FROM `dice_issues`.*").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "type", "assignee", "title", "creator"}).
			AddRow(uint64(1), uint64(1), "TASK", "2", "issue-title", "1"))
	mock.ExpectQuery("SELECT .* FROM `dice_issues`.*").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "type", "assignee", "title", "creator"}).
			AddRow(uint64(1), uint64(1), "TASK", "2", "issue-title", "1"))
	mock.ExpectQuery("SELECT .* FROM `erda_issue_subscriber`.*").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"issue_id", "user_id"}).AddRow(int64(1), "2"))
	mock.ExpectQuery("SELECT .*dice_issue_relation.*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "name", "belong"}).
			AddRow(uint64(1), "TASK", "parent", "belong"))
	mock.ExpectClose()
	mock.MatchExpectationsInOrder(false)

	// Patch bundle methods with gomonkey to avoid network calls
	bundleType := reflect.TypeOf((*bundle.Bundle)(nil))
	patches := gomonkey.NewPatches()
	patches.ApplyFunc(discover.GetEndpoint, func(string) (string, error) { return "http://localhost:9093", nil })
	patches.ApplyMethod(bundleType, "GetProjectWithSetter",
		func(_ *bundle.Bundle, _ uint64, _ ...httpclient.RequestSetter) (*apistructs.ProjectDTO, error) {
			return &apistructs.ProjectDTO{ID: 1, OrgID: 1, Name: "project"}, nil
		},
	)
	patches.ApplyMethod(bundleType, "GetCurrentUser",
		func(_ *bundle.Bundle, _ string) (*apistructs.UserInfo, error) {
			return &apistructs.UserInfo{ID: "2", Nick: "tester"}, nil
		},
	)
	patches.ApplyMethod(bundleType, "CreateEvent",
		func(_ *bundle.Bundle, _ *apistructs.EventCreateRequest) error { return nil },
	)
	defer patches.Reset()

	bdl := bundle.New()
	p := &provider{db: db, bdl: bdl, Org: EventOrgMock{}, commonTran: &mockTranslator{}}
	err := p.CreateIssueEvent(&common.IssueStreamCreateRequest{
		IssueID: 1,
	})
	assert.NoError(t, err)
	err = p.CreateIssueEvent(&common.IssueStreamCreateRequest{
		IssueID:    1,
		StreamType: common.ISTChangeContent,
	})
	assert.NoError(t, err)
}
