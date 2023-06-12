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

package issueexcel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
)

func Test_sortRelationsIntoBelongs(t *testing.T) {
	openA := &pb.IssueStateRelation{
		StateBelong: pb.IssueStateBelongEnum_StateBelong_name[int32(pb.IssueStateBelongEnum_OPEN)],
	}
	openB := &pb.IssueStateRelation{
		StateBelong: pb.IssueStateBelongEnum_StateBelong_name[int32(pb.IssueStateBelongEnum_OPEN)],
	}
	workingA := &pb.IssueStateRelation{
		StateBelong: pb.IssueStateBelongEnum_StateBelong_name[int32(pb.IssueStateBelongEnum_WORKING)],
	}
	doneA := &pb.IssueStateRelation{
		StateBelong: pb.IssueStateBelongEnum_StateBelong_name[int32(pb.IssueStateBelongEnum_DONE)],
	}
	newWorkingB := &pb.IssueStateRelation{
		StateBelong: pb.IssueStateBelongEnum_StateBelong_name[int32(pb.IssueStateBelongEnum_WORKING)],
	}
	newDoneB := &pb.IssueStateRelation{
		StateBelong: pb.IssueStateBelongEnum_StateBelong_name[int32(pb.IssueStateBelongEnum_DONE)],
	}
	relations := []*pb.IssueStateRelation{openA, openB, workingA, doneA, newWorkingB, newDoneB}
	sortRelationsIntoBelongs("", relations)
	assert.Equal(t, []*pb.IssueStateRelation{openA, openB, workingA, newWorkingB, doneA, newDoneB}, relations)
}

// Test_tryToGuessNewStateBelong
// see rule at method comment: tryToGuessNewStateBelong
func Test_tryToGuessNewStateBelong(t *testing.T) {
	now := time.Now()

	assert.Equal(t, pb.IssueStateBelongEnum_CLOSED, tryToGuessNewStateBelong("任意状态", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_BUG, FinishAt: &now}}))
	assert.Equal(t, pb.IssueStateBelongEnum_DONE, tryToGuessNewStateBelong("任意状态", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_TASK, FinishAt: &now}}))

	assert.Equal(t, pb.IssueStateBelongEnum_CLOSED, tryToGuessNewStateBelong("已", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_BUG}}))
	assert.Equal(t, pb.IssueStateBelongEnum_DONE, tryToGuessNewStateBelong("已", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_TASK}}))

	assert.Equal(t, pb.IssueStateBelongEnum_OPEN, tryToGuessNewStateBelong("未", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_BUG}}))
	assert.Equal(t, pb.IssueStateBelongEnum_OPEN, tryToGuessNewStateBelong("待", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_BUG}}))
	assert.Equal(t, pb.IssueStateBelongEnum_OPEN, tryToGuessNewStateBelong("未", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_TASK}}))
	assert.Equal(t, pb.IssueStateBelongEnum_OPEN, tryToGuessNewStateBelong("待", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_TASK}}))

	assert.Equal(t, pb.IssueStateBelongEnum_WORKING, tryToGuessNewStateBelong("中", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_BUG}}))
	assert.Equal(t, pb.IssueStateBelongEnum_WORKING, tryToGuessNewStateBelong("正在", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_BUG}}))
	assert.Equal(t, pb.IssueStateBelongEnum_WORKING, tryToGuessNewStateBelong("中", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_TASK}}))
	assert.Equal(t, pb.IssueStateBelongEnum_WORKING, tryToGuessNewStateBelong("正在", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_TASK}}))

	assert.Equal(t, pb.IssueStateBelongEnum_CLOSED, tryToGuessNewStateBelong("完成", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_BUG}}))
	assert.Equal(t, pb.IssueStateBelongEnum_CLOSED, tryToGuessNewStateBelong("关闭", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_BUG}}))
	assert.Equal(t, pb.IssueStateBelongEnum_DONE, tryToGuessNewStateBelong("完成", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_TASK}}))
	assert.Equal(t, pb.IssueStateBelongEnum_DONE, tryToGuessNewStateBelong("关闭", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_TASK}}))

	assert.Equal(t, pb.IssueStateBelongEnum_OPEN, tryToGuessNewStateBelong("新建", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_BUG}}))
	assert.Equal(t, pb.IssueStateBelongEnum_OPEN, tryToGuessNewStateBelong("新建", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_TASK}}))

	assert.Equal(t, pb.IssueStateBelongEnum_WORKING, tryToGuessNewStateBelong("任意状态", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_BUG}}))
	assert.Equal(t, pb.IssueStateBelongEnum_WORKING, tryToGuessNewStateBelong("任意状态", IssueSheetModel{Common: IssueSheetModelCommon{IssueType: pb.IssueTypeEnum_TASK}}))
}
