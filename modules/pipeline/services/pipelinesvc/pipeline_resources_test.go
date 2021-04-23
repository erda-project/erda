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

package pipelinesvc

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func genTaskWithLimitCPUMem(limitCPU, limitMem float64) *spec.PipelineTask {
	return &spec.PipelineTask{
		Extra: spec.PipelineTaskExtra{
			AppliedResources: apistructs.PipelineAppliedResources{
				Limits: apistructs.PipelineAppliedResource{
					CPU:      limitCPU,
					MemoryMB: limitMem,
				},
			},
		},
	}
}

func genTaskWithRequestCPUMem(limitCPU, limitMem float64) *spec.PipelineTask {
	return &spec.PipelineTask{
		Extra: spec.PipelineTaskExtra{
			AppliedResources: apistructs.PipelineAppliedResources{
				Requests: apistructs.PipelineAppliedResource{
					CPU:      limitCPU,
					MemoryMB: limitMem,
				},
			},
		},
	}
}

func Test_calculatePipelineLimitResource(t *testing.T) {
	type args struct {
		allStagedTasks [][]*spec.PipelineTask
	}
	tests := []struct {
		name string
		args args
		want apistructs.PipelineAppliedResource
	}{
		{
			// cpu
			// 1 2 (3)
			// 2 3 (5)
			// 4   (4)
			// => max((1+2), (2+3), (4)) = 5
			// mem
			// 5 4 (9)
			// 1 3 (4)
			// 8   (8)
			// => max((5+4), (1+3), (8)) = 9
			name: "enough cpu",
			args: args{
				allStagedTasks: [][]*spec.PipelineTask{
					{genTaskWithLimitCPUMem(1, 5), genTaskWithLimitCPUMem(2, 4)},
					{genTaskWithLimitCPUMem(2, 1), genTaskWithLimitCPUMem(3, 3)},
					{genTaskWithLimitCPUMem(4, 8)},
				},
			},
			want: apistructs.PipelineAppliedResource{
				CPU:      5,
				MemoryMB: 9,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculatePipelineLimitResource(tt.args.allStagedTasks); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculatePipelineLimitResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculatePipelineRequestResource(t *testing.T) {
	type args struct {
		allStagedTasks [][]*spec.PipelineTask
	}
	tests := []struct {
		name string
		args args
		want apistructs.PipelineAppliedResource
	}{
		{
			// calculate minResource
			// cpu
			// 1 2 (2)
			// 2 3 (3)
			// 4   (4)
			// => max(1,2,2,3,4) = 4
			// mem
			// 5 4 (5)
			// 4 1 (4)
			// 7   (7)
			// => max(5,4,4,1,7) = 7
			name: "minimal resource",
			args: args{
				allStagedTasks: [][]*spec.PipelineTask{
					{genTaskWithRequestCPUMem(1, 5), genTaskWithRequestCPUMem(2, 4)},
					{genTaskWithRequestCPUMem(2, 4), genTaskWithRequestCPUMem(3, 1)},
					{genTaskWithRequestCPUMem(4, 7)},
				},
			},
			want: apistructs.PipelineAppliedResource{
				CPU:      4,
				MemoryMB: 7,
			},
		},
		{
			name: "test",
			args: args{
				allStagedTasks: [][]*spec.PipelineTask{
					{genTaskWithRequestCPUMem(1, 1024), genTaskWithRequestCPUMem(1, 1024)},
					{genTaskWithRequestCPUMem(2, 2049)},
				},
			},
			want: apistructs.PipelineAppliedResource{
				CPU:      2,
				MemoryMB: 2049,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculatePipelineRequestResource(tt.args.allStagedTasks); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculatePipelineRequestResource() = %v, want %v", got, tt.want)
			}
		})
	}
}
