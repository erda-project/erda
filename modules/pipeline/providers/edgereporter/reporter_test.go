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

package edgereporter

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/modules/pipeline/spec"
)

func Test_pipelineFilterIn(t *testing.T) {
	type args struct {
		pipelines []spec.PipelineBase
		fn        func(p *spec.PipelineBase) bool
	}
	tests := []struct {
		name string
		args args
		want []spec.PipelineBase
	}{
		{
			name: "",
			args: args{
				pipelines: []spec.PipelineBase{
					{
						ID:     1,
						Status: "Success",
					},
					{
						ID:     2,
						Status: "Running",
					},
					{
						ID:     3,
						Status: "Analyzed",
					},
					{
						ID:     4,
						Status: "Failed",
					},
				},
				fn: func(p *spec.PipelineBase) bool {
					return p.Status.IsEndStatus()
				},
			},
			want: []spec.PipelineBase{
				{
					ID:     1,
					Status: "Success",
				},
				{
					ID:     4,
					Status: "Failed",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pipelineFilterIn(tt.args.pipelines, tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pipelineFilterIn() = %v, want %v", got, tt.want)
			}
		})
	}
}
