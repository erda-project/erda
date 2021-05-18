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

package statusutil

import (
	"fmt"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func TestCalculatePipelineStatusV2(t *testing.T) {
	var datas = []struct {
		tasks        []*spec.PipelineTask
		expectStatus apistructs.PipelineStatus
		desc         string
	}{
		{
			tasks: []*spec.PipelineTask{
				{
					Status: apistructs.PipelineStatusAnalyzed,
				},
			},
			expectStatus: apistructs.PipelineStatusAnalyzed,
			desc:         "展示分析完成",
		},
		{
			tasks: []*spec.PipelineTask{
				{
					Status: apistructs.PipelineStatusAnalyzed,
				},
				{
					Status: apistructs.PipelineStatusRunning,
				},
			},
			expectStatus: apistructs.PipelineStatusRunning,
			desc:         "展示运行中",
		},
		{
			tasks: []*spec.PipelineTask{
				{
					Status: apistructs.PipelineStatusAnalyzed,
				},
				{
					Status: apistructs.PipelineStatusAnalyzeFailed,
				},
			},
			expectStatus: apistructs.PipelineStatusFailed,
			desc:         "有一个失败的状态，应该展示失败",
		},
		{
			tasks: []*spec.PipelineTask{
				{
					Status: apistructs.PipelineStatusAnalyzed,
				},
				{
					Status: apistructs.PipelineStatusBorn,
				},
				{
					Status: apistructs.PipelineStatusSuccess,
				},
			},
			expectStatus: apistructs.PipelineStatusRunning,
			desc:         "有成功有运行中，应该展示运行中",
		},
		{
			tasks: []*spec.PipelineTask{
				{
					Status: apistructs.PipelineStatusBorn,
				},
				{
					Status: apistructs.PipelineStatusPaused,
				},
				{
					Status: apistructs.PipelineStatusSuccess,
				},
			},
			expectStatus: apistructs.PipelineStatusPaused,
			desc:         "有暂停有准备中，展示暂停",
		},
		{
			tasks: []*spec.PipelineTask{
				{
					Status: apistructs.PipelineStatusBorn,
				},
				{
					Status: apistructs.PipelineStatusPaused,
				},
				{
					Status: apistructs.PipelineStatusQueue,
				},
			},
			expectStatus: apistructs.PipelineStatusRunning,
			desc:         "有排队有等待有准备中，展示运行中",
		},
	}

	for index, v := range datas {
		pipelineStatus := CalculatePipelineStatusV2(v.tasks)
		if pipelineStatus != v.expectStatus {
			fmt.Println(fmt.Sprintf("[%v] desc %s. %s not equal to %s", index, v.desc, pipelineStatus, v.expectStatus))
			t.Fail()
		}
	}
}

func TestCalculatePipelineTaskAllDone(t *testing.T) {
	var datas = []struct {
		tasks         []*spec.PipelineTask
		expectAllDone bool
		desc          string
	}{
		{
			tasks: []*spec.PipelineTask{
				{Status: apistructs.PipelineStatusAnalyzed},
			},
			expectAllDone: false,
			desc:          "展示分析完成",
		},
		{
			tasks: []*spec.PipelineTask{
				{Status: apistructs.PipelineStatusAnalyzed},
				{Status: apistructs.PipelineStatusRunning},
			},
			expectAllDone: false,
			desc:          "展示运行中",
		},
		{
			tasks: []*spec.PipelineTask{
				{Status: apistructs.PipelineStatusAnalyzed},
				{Status: apistructs.PipelineStatusAnalyzeFailed},
			},
			expectAllDone: false,
			desc:          "有一个失败的状态，应该展示失败",
		},
		{
			tasks: []*spec.PipelineTask{
				{Status: apistructs.PipelineStatusSuccess},
				{Status: apistructs.PipelineStatusAnalyzeFailed},
				{Status: apistructs.PipelineStatusFailed},
			},
			expectAllDone: true,
			desc:          "所有都是终态",
		},
		{
			tasks: []*spec.PipelineTask{
				{Status: apistructs.PipelineStatusAnalyzed},
				{Status: apistructs.PipelineStatusBorn},
				{Status: apistructs.PipelineStatusSuccess},
			},
			expectAllDone: false,
			desc:          "有成功有运行中，应该展示运行中",
		},
		{
			tasks: []*spec.PipelineTask{
				{Status: apistructs.PipelineStatusBorn},
				{Status: apistructs.PipelineStatusPaused},
				{Status: apistructs.PipelineStatusSuccess},
			},
			expectAllDone: false,
			desc:          "有暂停有准备中，展示暂停",
		},
		{
			tasks: []*spec.PipelineTask{
				{Status: apistructs.PipelineStatusBorn},
				{Status: apistructs.PipelineStatusPaused},
				{Status: apistructs.PipelineStatusQueue},
			},
			expectAllDone: false,
			desc:          "有排队有等待有准备中，展示运行中",
		},
		{
			tasks: []*spec.PipelineTask{
				{Status: apistructs.PipelineStatusSuccess},
				{Status: apistructs.PipelineStatusFailed},
				{Status: apistructs.PipelineStatusDisabled},
			},
			expectAllDone: true,
			desc:          "成功 + 失败 + 禁用 = 终态",
		},
	}

	for index, v := range datas {
		allDone := CalculatePipelineTaskAllDone(v.tasks)
		if allDone != v.expectAllDone {
			fmt.Println(fmt.Sprintf("[%d] desc %s. %v (actual) not equal to %v (expected)", index, v.desc, allDone, v.expectAllDone))
			t.Fail()
		}
	}
}
