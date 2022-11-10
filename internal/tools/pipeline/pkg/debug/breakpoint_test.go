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

package debug

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
)

func TestMergeBreakpoint(t *testing.T) {
	type args struct {
		taskConfig     *pb.Breakpoint
		pipelineConfig *pb.Breakpoint
	}
	tests := []struct {
		name      string
		args      args
		wantDebug bool
	}{
		{
			name: "pipeline want debug, task not want debug",
			args: args{
				taskConfig: &pb.Breakpoint{
					On: &pb.BreakpointOn{
						Failure: false,
					},
				},
				pipelineConfig: &pb.Breakpoint{
					On: &pb.BreakpointOn{
						Failure: true,
					},
				},
			},
			wantDebug: false,
		},
		{
			name: "pipeline not want debug, task want debug",
			args: args{
				taskConfig: &pb.Breakpoint{
					On: &pb.BreakpointOn{
						Failure: true,
					},
				},
				pipelineConfig: &pb.Breakpoint{
					On: &pb.BreakpointOn{
						Failure: false,
					},
				},
			},
			wantDebug: true,
		},
		{
			name: "both want debug",
			args: args{
				taskConfig: &pb.Breakpoint{
					On: &pb.BreakpointOn{
						Failure: true,
					},
				},
				pipelineConfig: &pb.Breakpoint{
					On: &pb.BreakpointOn{
						Failure: true,
					},
				},
			},
			wantDebug: true,
		},
		{
			name: "both not want debug",
			args: args{
				taskConfig: &pb.Breakpoint{
					On: &pb.BreakpointOn{
						Failure: false,
					},
				},
				pipelineConfig: &pb.Breakpoint{
					On: &pb.BreakpointOn{
						Failure: false,
					},
				},
			},
			wantDebug: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeBreakpoint(tt.args.taskConfig, tt.args.pipelineConfig)
			assert.Equal(t, tt.wantDebug, got.On.Failure)
		})
	}
}

func TestParseDebugTimeout(t *testing.T) {
	type args struct {
		timeout *structpb.Value
	}
	tests := []struct {
		name    string
		args    args
		want    time.Duration
		wantErr bool
	}{
		{
			name: "timeout is 1m",
			args: args{
				timeout: structpb.NewStringValue("1m"),
			},
			want:    time.Duration(60) * time.Second,
			wantErr: false,
		},
		{
			name: "timeout is 1h",
			args: args{
				timeout: structpb.NewStringValue("1h"),
			},
			want:    time.Duration(3600) * time.Second,
			wantErr: false,
		},
		{
			name: "360 second use float64",
			args: args{
				timeout: structpb.NewNumberValue(360),
			},
			want:    time.Duration(360) * time.Second,
			wantErr: false,
		},
		{
			name: "invalid string timeout",
			args: args{
				timeout: structpb.NewStringValue("22od"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDebugTimeout(tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDebugTimeout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want, *got)
			}
		})
	}
}
