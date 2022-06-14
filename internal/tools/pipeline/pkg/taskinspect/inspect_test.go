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

package taskinspect

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/tools/pipeline/pkg/taskerror"
)

func TestIsErrorsExceed(t *testing.T) {
	tests := []struct {
		name    string
		inspect *PipelineTaskInspect
		want    bool
	}{
		{
			name: "less than one hour and count less than 180",
			inspect: &PipelineTaskInspect{
				Errors: []*taskerror.PipelineTaskErrResponse{&taskerror.PipelineTaskErrResponse{Msg: "xxx", Ctx: taskerror.PipelineTaskErrCtx{StartTime: time.Now().Add(-59 * time.Minute), Count: 179, EndTime: time.Now()}}},
			},
			want: false,
		},
		{
			name: "less than one hour but count more than 180",
			inspect: &PipelineTaskInspect{
				Errors: []*taskerror.PipelineTaskErrResponse{&taskerror.PipelineTaskErrResponse{Msg: "xxx", Ctx: taskerror.PipelineTaskErrCtx{StartTime: time.Now().Add(-59 * time.Minute), Count: 181, EndTime: time.Now()}}},
			},
			want: true,
		},
		{
			name: "more than one hour ans count less than 180 per hour",
			inspect: &PipelineTaskInspect{
				Errors: []*taskerror.PipelineTaskErrResponse{&taskerror.PipelineTaskErrResponse{Msg: "xxx", Ctx: taskerror.PipelineTaskErrCtx{StartTime: time.Now().Add(-61 * time.Minute), Count: 180, EndTime: time.Now()}}},
			},
			want: false,
		},
		{
			name: "more than one hour ans count more than 180 per hour",
			inspect: &PipelineTaskInspect{
				Errors: []*taskerror.PipelineTaskErrResponse{&taskerror.PipelineTaskErrResponse{Msg: "xxx", Ctx: taskerror.PipelineTaskErrCtx{StartTime: time.Now().Add(-61 * time.Minute), Count: 185, EndTime: time.Now()}}},
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

func TestPipelineTaskInspect_ConvertErrors(t1 *testing.T) {
	type fields struct {
		Inspect     string
		Events      string
		MachineStat *PipelineTaskMachineStat
		Errors      []*taskerror.PipelineTaskErrResponse
	}
	tests := []struct {
		name      string
		fields    fields
		converted bool
	}{
		{
			name: "count = 1",
			fields: fields{
				Errors: []*taskerror.PipelineTaskErrResponse{
					{
						Msg: "count = 1",
						Ctx: taskerror.PipelineTaskErrCtx{
							Count: 1,
						},
					},
				},
			},
			converted: false,
		},
		{
			name: "count = 2",
			fields: fields{
				Errors: []*taskerror.PipelineTaskErrResponse{
					{
						Msg: "count = 2",
						Ctx: taskerror.PipelineTaskErrCtx{
							Count: 2,
						},
					},
				},
			},
			converted: true,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &PipelineTaskInspect{
				Inspect:     tt.fields.Inspect,
				Events:      tt.fields.Events,
				MachineStat: tt.fields.MachineStat,
				Errors:      tt.fields.Errors,
			}
			t.ConvertErrors()
			resp := t.Errors[0]
			assert.Equal(t1, tt.converted, strings.Contains(resp.Msg, "startTime: "))
		})
	}
}
