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

package apistructs

import (
	"fmt"
	"testing"
	"time"
)

func TestPipelineTaskLoop_Duplicate(t *testing.T) {
	var l *PipelineTaskLoop
	fmt.Println(l.Duplicate())
}

func TestPipelineTaskLoop_IsEmpty(t *testing.T) {
	type fields struct {
		Break    string
		Strategy *LoopStrategy
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "test_break",
			fields: fields{
				Break: "test",
			},
			want: false,
		},
		{
			name: "test_DeclineLimitSec",
			fields: fields{
				Strategy: &LoopStrategy{
					DeclineLimitSec: 1,
				},
			},
			want: false,
		},
		{
			name: "test_DeclineRatio",
			fields: fields{
				Strategy: &LoopStrategy{
					DeclineLimitSec: 1,
				},
			},
			want: false,
		},
		{
			name: "test_IntervalSec",
			fields: fields{
				Strategy: &LoopStrategy{
					DeclineLimitSec: 1,
				},
			},
			want: false,
		},
		{
			name: "test_MaxTimes",
			fields: fields{
				Strategy: &LoopStrategy{
					DeclineLimitSec: 1,
				},
			},
			want: false,
		},
		{
			name:   "test_MaxTimes",
			fields: fields{},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &PipelineTaskLoop{
				Break:    tt.fields.Break,
				Strategy: tt.fields.Strategy,
			}
			if got := l.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsErrorsExceed(t *testing.T) {
	tests := []struct {
		name    string
		inspect *PipelineTaskInspect
		want    bool
	}{
		{
			name: "less than one hour and count less than 180",
			inspect: &PipelineTaskInspect{
				Errors: []*PipelineTaskErrResponse{&PipelineTaskErrResponse{Msg: "xxx", Ctx: PipelineTaskErrCtx{StartTime: time.Now().Add(-59 * time.Minute), Count: 179, EndTime: time.Now()}}},
			},
			want: false,
		},
		{
			name: "less than one hour but count more than 180",
			inspect: &PipelineTaskInspect{
				Errors: []*PipelineTaskErrResponse{&PipelineTaskErrResponse{Msg: "xxx", Ctx: PipelineTaskErrCtx{StartTime: time.Now().Add(-59 * time.Minute), Count: 181, EndTime: time.Now()}}},
			},
			want: true,
		},
		{
			name: "more than one hour ans count less than 180 per hour",
			inspect: &PipelineTaskInspect{
				Errors: []*PipelineTaskErrResponse{&PipelineTaskErrResponse{Msg: "xxx", Ctx: PipelineTaskErrCtx{StartTime: time.Now().Add(-61 * time.Minute), Count: 180, EndTime: time.Now()}}},
			},
			want: false,
		},
		{
			name: "more than one hour ans count more than 180 per hour",
			inspect: &PipelineTaskInspect{
				Errors: []*PipelineTaskErrResponse{&PipelineTaskErrResponse{Msg: "xxx", Ctx: PipelineTaskErrCtx{StartTime: time.Now().Add(-61 * time.Minute), Count: 185, EndTime: time.Now()}}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		got, _ := tt.inspect.IsErrorsExceed()
		if got != tt.want {
			t.Errorf("%s want: %v, but got: %v", tt.name, tt.want, got)
		}
	}
}
