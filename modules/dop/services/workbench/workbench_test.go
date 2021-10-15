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

package workbench

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/issue"
)

func TestSetDiffFinishedIssueNum(t *testing.T) {
	issueSvc := &issue.Issue{}
	m := monkey.PatchInstanceMethod(reflect.TypeOf(issueSvc), "GetIssueNumByPros",
		func(issueSvc *issue.Issue, projectIDS []uint64, req apistructs.IssuePagingRequest) ([]apistructs.IssueNum, error) {
			res := []apistructs.IssueNum{}
			for _, proID := range projectIDS {
				num := apistructs.IssueNum{
					ProjectID: proID,
					IssueNum:  3,
				}
				res = append(res, num)
			}
			return res, nil
		})
	defer m.Unpatch()

	workbench := New(WithIssue(issueSvc))
	items := []*apistructs.WorkbenchProjectItem{
		&apistructs.WorkbenchProjectItem{
			ProjectDTO: apistructs.ProjectDTO{ID: 1},
		},
		&apistructs.WorkbenchProjectItem{
			ProjectDTO: apistructs.ProjectDTO{ID: 2},
		},
	}
	req := apistructs.IssuePagingRequest{
		IssueListRequest: apistructs.IssueListRequest{
			State: []int64{1, 2, 3},
		},
	}
	err := workbench.SetDiffFinishedIssueNum(req, items)
	assert.NoError(t, err)
	assert.Equal(t, 3, items[0].ExpiredIssueNum)
	assert.Equal(t, 3, items[1].UnSpecialIssueNum)
}

func TestWorkbench_GetUndoneProjectItems(t *testing.T) {
	issueSvc := &issue.Issue{}
	item := &apistructs.WorkbenchProjectItem{
		TotalIssueNum: 1,
		IssueList: []apistructs.Issue{
			{
				ID: 1,
			},
		},
	}
	m := monkey.PatchInstanceMethod(reflect.TypeOf(issueSvc), "GetIssuesByStates",
		func(issueSvc *issue.Issue, req apistructs.WorkbenchRequest) (map[uint64]*apistructs.WorkbenchProjectItem, error) {
			return map[uint64]*apistructs.WorkbenchProjectItem{
				1: item,
			}, nil
		})
	defer m.Unpatch()

	workbench := New(WithIssue(issueSvc))

	type args struct {
		req    apistructs.WorkbenchRequest
		userID string
	}
	tests := []struct {
		name    string
		args    args
		want    *apistructs.WorkbenchResponse
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				req: apistructs.WorkbenchRequest{},
			},
			want: &apistructs.WorkbenchResponse{
				Data: apistructs.WorkbenchResponseData{
					TotalProject: 1,
					List: []*apistructs.WorkbenchProjectItem{
						item,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := workbench.GetUndoneProjectItems(tt.args.req, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Workbench.GetUndoneProjectItems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Workbench.GetUndoneProjectItems() = %v, want %v", got, tt.want)
			}
		})
	}
}
