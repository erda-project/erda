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

package pipelinesvc

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
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

func TestPipelineSvc_judgeTaskExecutor(t *testing.T) {
	type args struct {
		action     *pipelineyml.Action
		actionSpec *apistructs.ActionSpec
	}
	tests := []struct {
		name    string
		args    args
		want    spec.PipelineTaskExecutorKind
		want1   spec.PipelineTaskExecutorName
		wantErr bool
	}{
		{
			name: "empty executor",
			args: args{
				actionSpec: &apistructs.ActionSpec{
					Executor: nil,
				},
			},
			want:    spec.PipelineTaskExecutorKindScheduler,
			want1:   spec.PipelineTaskExecutorNameSchedulerDefault,
			wantErr: false,
		},
		{
			name:    "empty spec",
			args:    args{},
			want:    spec.PipelineTaskExecutorKindScheduler,
			want1:   spec.PipelineTaskExecutorNameSchedulerDefault,
			wantErr: false,
		},
		{
			name: "not match kind",
			args: args{
				actionSpec: &apistructs.ActionSpec{
					Executor: &apistructs.ActionExecutor{
						Name: spec.PipelineTaskExecutorNameEmpty.String(),
						Kind: "__other",
					},
				},
			},
			want:    spec.PipelineTaskExecutorKindScheduler,
			want1:   spec.PipelineTaskExecutorNameSchedulerDefault,
			wantErr: false,
		},
		{
			name: "not match name",
			args: args{
				actionSpec: &apistructs.ActionSpec{
					Executor: &apistructs.ActionExecutor{
						Kind: string(spec.PipelineTaskExecutorKindMemory),
						Name: "__other",
					},
				},
			},
			want:    spec.PipelineTaskExecutorKindScheduler,
			want1:   spec.PipelineTaskExecutorNameSchedulerDefault,
			wantErr: false,
		},
		{
			name: "normal",
			args: args{
				actionSpec: &apistructs.ActionSpec{
					Executor: &apistructs.ActionExecutor{
						Kind: string(spec.PipelineTaskExecutorKindAPITest),
						Name: spec.PipelineTaskExecutorNameSchedulerDefault.String(),
					},
				},
			},
			want:    spec.PipelineTaskExecutorKindAPITest,
			want1:   spec.PipelineTaskExecutorNameSchedulerDefault,
			wantErr: false,
		},
		{
			name: "not find kind or name",
			args: args{
				actionSpec: &apistructs.ActionSpec{
					Executor: &apistructs.ActionExecutor{
						Kind: "__test_kind",
						Name: "__test_name",
					},
				},
			},
			want:    spec.PipelineTaskExecutorKindScheduler,
			want1:   spec.PipelineTaskExecutorNameSchedulerDefault,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PipelineSvc{}
			got, got1, err := s.judgeTaskExecutor(tt.args.action, tt.args.actionSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("judgeTaskExecutor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("judgeTaskExecutor() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("judgeTaskExecutor() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
