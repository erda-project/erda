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

func TestSetSpecialIssueNum(t *testing.T) {
	issueSvc := &issue.Issue{}
	m := monkey.PatchInstanceMethod(reflect.TypeOf(issueSvc), "Paging",
		func(issueSvc *issue.Issue, req apistructs.IssuePagingRequest) ([]apistructs.Issue, uint64, error) {
			return []apistructs.Issue{}, 3, nil
		})
	defer m.Unpatch()

	workbench := New(WithIssue(issueSvc))
	item := apistructs.WorkbenchProjectItem{}
	err := workbench.SetSpecialIssueNum("1", &item)
	assert.NoError(t, err)
	assert.Equal(t, 3, item.ExpiredIssueNum)
	assert.Equal(t, 3, item.FeatureDayNum)
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
	assert.Equal(t, 5, item.ExpiredOneDayNum, 5)
}
