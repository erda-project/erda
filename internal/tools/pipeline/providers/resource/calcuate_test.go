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

package resource

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func Test_calculatePipelineLimitResource(t *testing.T) {
	type args struct {
		allStagedTasksResources [][]*apistructs.PipelineAppliedResources
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
				allStagedTasksResources: [][]*apistructs.PipelineAppliedResources{
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
			if got := calculatePipelineLimitResource(tt.args.allStagedTasksResources); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculatePipelineLimitResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculatePipelineRequestResource(t *testing.T) {
	type args struct {
		allStagedTasksResources [][]*apistructs.PipelineAppliedResources
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
				allStagedTasksResources: [][]*apistructs.PipelineAppliedResources{
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
				allStagedTasksResources: [][]*apistructs.PipelineAppliedResources{
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
			if got := calculatePipelineRequestResource(tt.args.allStagedTasksResources); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculatePipelineRequestResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func genTaskWithLimitCPUMem(limitCPU, limitMem float64) *apistructs.PipelineAppliedResources {
	return &apistructs.PipelineAppliedResources{
		Limits: apistructs.PipelineAppliedResource{
			CPU:      limitCPU,
			MemoryMB: limitMem,
		},
	}
}

func genTaskWithRequestCPUMem(limitCPU, limitMem float64) *apistructs.PipelineAppliedResources {
	return &apistructs.PipelineAppliedResources{
		Requests: apistructs.PipelineAppliedResource{
			CPU:      limitCPU,
			MemoryMB: limitMem,
		},
	}
}

func Test_calculateNormalTaskLimitResource(t *testing.T) {
	type args struct {
		action       *pipelineyml.Action
		actionDefine *diceyml.Job
		defaultRes   apistructs.PipelineAppliedResource
	}
	tests := []struct {
		name string
		args args
		want apistructs.PipelineAppliedResource
	}{
		{
			name: "all defined",
			args: args{
				action:       genActionWithRes(1, 0, 2049),
				actionDefine: genActionDefineWithRes(2, 3, 1024, 2048),
				defaultRes:   apistructs.PipelineAppliedResource{CPU: 0.1, MemoryMB: 32},
			},
			want: apistructs.PipelineAppliedResource{
				CPU:      3,
				MemoryMB: 2049,
			},
		},
		{
			name: "no custom action resources defined",
			args: args{
				action:       genActionWithRes(0, 0, 0),
				actionDefine: genActionDefineWithRes(2, 3, 1024, 2048),
				defaultRes:   apistructs.PipelineAppliedResource{CPU: 0.1, MemoryMB: 32},
			},
			want: apistructs.PipelineAppliedResource{
				CPU:      3,
				MemoryMB: 2048,
			},
		},
		{
			name: "all undefined",
			args: args{
				action:       genActionWithRes(0, 0, 0),
				actionDefine: genActionDefineWithRes(0, 0, 0, 0),
				defaultRes:   apistructs.PipelineAppliedResource{CPU: 0.1, MemoryMB: 32},
			},
			want: apistructs.PipelineAppliedResource{
				CPU:      0.1,
				MemoryMB: 32,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateNormalTaskLimitResource(tt.args.action, tt.args.actionDefine, tt.args.defaultRes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateNormalTaskLimitResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateOversoldTaskLimitResource(t *testing.T) {
	type args struct {
		action       *pipelineyml.Action
		actionDefine *diceyml.Job
		defaultRes   apistructs.PipelineAppliedResource
	}
	tests := []struct {
		name string
		args args
		want apistructs.PipelineAppliedResource
	}{
		{
			name: "cpu lower than define and default",
			args: args{
				action:       genActionWithRes(0.1, 0, 2049),
				actionDefine: genActionDefineWithRes(0.2, 0.4, 1024, 2048),
				defaultRes:   apistructs.PipelineAppliedResource{CPU: 0.5, MemoryMB: 32},
			},
			want: apistructs.PipelineAppliedResource{
				CPU:      0.8,
				MemoryMB: 2049,
			},
		},
		{
			name: "cpu bigger than define but lower than default",
			args: args{
				action:       genActionWithRes(0.2, 0, 0),
				actionDefine: genActionDefineWithRes(0.1, 0.5, 1024, 2048),
				defaultRes:   apistructs.PipelineAppliedResource{CPU: 1, MemoryMB: 32},
			},
			want: apistructs.PipelineAppliedResource{
				CPU:      1,
				MemoryMB: 2048,
			},
		},
		{
			name: "cpu bigger than define and default",
			args: args{
				action:       genActionWithRes(3, 0, 0),
				actionDefine: genActionDefineWithRes(2, 0, 0, 0),
				defaultRes:   apistructs.PipelineAppliedResource{CPU: 1, MemoryMB: 32},
			},
			want: apistructs.PipelineAppliedResource{
				CPU:      2,
				MemoryMB: 32,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateOversoldTaskLimitResource(calculateNormalTaskLimitResource(tt.args.action, tt.args.actionDefine, tt.args.defaultRes), apistructs.PipelineOverSoldResource{
				CPURate: 2,
				MaxCPU:  2,
			}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateNormalTaskLimitResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateNormalTaskRequestResource(t *testing.T) {
	type args struct {
		action       *pipelineyml.Action
		actionDefine *diceyml.Job
		defaultRes   apistructs.PipelineAppliedResource
	}

	tests := []struct {
		name string
		args args
		want apistructs.PipelineAppliedResource
	}{
		{
			name: "all defined",
			args: args{
				action:       genActionWithRes(1, 0, 2049),
				actionDefine: genActionDefineWithRes(2, 3, 1024, 2048),
				defaultRes:   apistructs.PipelineAppliedResource{CPU: 0.1, MemoryMB: 32},
			},
			want: apistructs.PipelineAppliedResource{
				CPU:      1,
				MemoryMB: 2049,
			},
		},
		{
			name: "no custom action resources defined",
			args: args{
				action:       genActionWithRes(0, 0, 0),
				actionDefine: genActionDefineWithRes(2, 3, 1024, 2048),
				defaultRes:   apistructs.PipelineAppliedResource{CPU: 0.1, MemoryMB: 32},
			},
			want: apistructs.PipelineAppliedResource{
				CPU:      2,
				MemoryMB: 1024,
			},
		},
		{
			name: "all undefined",
			args: args{
				action:       genActionWithRes(0, 0, 0),
				actionDefine: genActionDefineWithRes(0, 0, 0, 0),
				defaultRes:   apistructs.PipelineAppliedResource{CPU: 0.1, MemoryMB: 32},
			},
			want: apistructs.PipelineAppliedResource{
				CPU:      0.1,
				MemoryMB: 32,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateNormalTaskRequestResource(tt.args.action, tt.args.actionDefine, tt.args.defaultRes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateNormalTaskRequestResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func genActionWithRes(cpu, maxCPU float64, memoryMB int) *pipelineyml.Action {
	return &pipelineyml.Action{
		Resources: pipelineyml.Resources{
			CPU:    cpu,
			MaxCPU: maxCPU,
			Mem:    memoryMB,
		},
	}
}

func genActionDefineWithRes(cpu, maxCPU float64, memoryMB, maxMemoryMB int) *diceyml.Job {
	return &diceyml.Job{
		Resources: diceyml.Resources{
			CPU:    cpu,
			MaxCPU: maxCPU,
			Mem:    memoryMB,
			MaxMem: maxMemoryMB,
		},
	}
}
