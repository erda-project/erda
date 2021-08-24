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

package bundle

//import (
//	"fmt"
//	"os"
//	"testing"
//
//	"github.com/erda-project/erda/apistructs"
//
//	"github.com/davecgh/go-spew/spew"
//	"github.com/stretchr/testify/assert"
//)
//
//func TestGetIssue(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
//	bdl := New(WithCMDB())
//
//	issue, err := bdl.GetIssue(74)
//	assert.NoError(t, err)
//
//	spew.Dump(issue)
//}
//
//func TestPageIssues(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
//	bdl := New(WithCMDB())
//	req := apistructs.IssuePagingRequest{
//		PageSize: 200,
//		IssueListRequest: apistructs.IssueListRequest{
//			Title:           "",
//			Type:            []apistructs.IssueType{apistructs.IssueTypeTask},
//			ProjectID:       2,
//			IterationID:     0,
//			IterationIDs:    nil,
//			AppID:           nil,
//			RequirementID:   nil,
//			Creators:        nil,
//			Assignees:       nil,
//			Label:           nil,
//			StartCreatedAt:  0,
//			EndCreatedAt:    0,
//			StartFinishedAt: 0,
//			EndFinishedAt:   0,
//			RelatedIssueIDs: nil,
//			Source:          "",
//			OrderBy:         "",
//			TaskType:        nil,
//			BugStage:        nil,
//			Owner:           nil,
//			Asc:             false,
//			IDs:             nil,
//			IdentityInfo: apistructs.IdentityInfo{
//				UserID: "2",
//			},
//			External: false,
//		},
//	}
//
//	is, err := bdl.PageIssues(req)
//	assert.NoError(t, err)
//	spew.Dump(is)
//}
//
//func TestGetIssueTypeState(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
//	bdl := New(WithCMDB())
//	req := apistructs.IssueStateRelationGetRequest{
//		ProjectID: 35,
//		IssueType: apistructs.IssueTypeTask,
//	}
//	rsp, err := bdl.GetIssueStateBelong(req)
//	assert.Equal(t, nil, err)
//	t.Logf("issue type state info:%v", rsp)
//}
//
//func TestGetIssueStage(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "localhost:9093")
//	bdl := New(WithCMDB())
//	stages, err := bdl.GetIssueStage(2, apistructs.IssueTypeBug)
//	assert.NoError(t, err)
//	fmt.Println(len(stages))
//}
