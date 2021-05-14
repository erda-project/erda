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
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

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
