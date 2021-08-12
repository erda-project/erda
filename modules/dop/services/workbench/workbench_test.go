// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
