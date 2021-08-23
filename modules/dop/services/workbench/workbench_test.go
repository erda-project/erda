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

func TestGetUndoneProjectItem(t *testing.T) {
	issueSvc := &issue.Issue{}
	m := monkey.PatchInstanceMethod(reflect.TypeOf(issueSvc), "Paging",
		func(issueSvc *issue.Issue, req apistructs.IssuePagingRequest) ([]apistructs.Issue, uint64, error) {
			return []apistructs.Issue{apistructs.Issue{ID: 1}}, 5, nil
		})
	defer m.Unpatch()

	workbench := New(WithIssue(issueSvc))
	item, err := workbench.GetUndoneProjectItem("1", 3, apistructs.ProjectDTO{})
	assert.NoError(t, err)
	assert.Equal(t, 5, item.TotalIssueNum)
}
