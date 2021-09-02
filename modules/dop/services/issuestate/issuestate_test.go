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

package issuestate

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
)

func TestIssueState_GetIssueStateIDs(t *testing.T) {
	is := New()
	db := &dao.DBClient{}
	m := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetIssuesStates",
		func(db *dao.DBClient, req *apistructs.IssueStatesGetRequest) ([]dao.IssueState, error) {
			return []dao.IssueState{
				{
					ProjectID: 1,
				},
			}, nil
		})
	defer m.Unpatch()

	res, err := is.GetIssueStateIDs(&apistructs.IssueStatesGetRequest{
		ProjectID: 1,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))
}
