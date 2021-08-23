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

package spec

import (
	"encoding/json"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

func TestRuntimeID(t *testing.T) {
	s := `{"metadata":[{"name":"runtimeID","value":"9","type":"link"},{"name":"operatorID","value":"2"}]}`
	r := apistructs.PipelineTaskResult{}
	if err := json.Unmarshal([]byte(s), &r); err != nil {
		logrus.Fatal(err)
	}
	pt := PipelineTask{Result: r}
	assert.Equal(t, pt.RuntimeID(), "9")
}

func TestTaskContextDedup(t *testing.T) {
	ctx := PipelineTaskContext{
		InStorages: apistructs.Metadata{
			{Name: "in1", Value: "v1"},
			{Name: "in2", Value: "v2"},
			{Name: "in1", Value: "v1_2"},
		},
		OutStorages: apistructs.Metadata{
			{Name: "out1", Value: "v1"},
			{Name: "out2", Value: "v2"},
			{Name: "out1", Value: "v1_2"},
		},
	}
	assert.Equal(t, len(ctx.InStorages), 3)
	assert.Equal(t, len(ctx.OutStorages), 3)

	ctx.Dedup()
	assert.Equal(t, len(ctx.InStorages), 2)
	assert.Equal(t, len(ctx.OutStorages), 2)
}

func TestPipelineTaskExecutorName_Check(t *testing.T) {
	tests := []struct {
		name string
		that PipelineTaskExecutorName
		want bool
	}{
		{
			name: "PipelineTaskExecutorNameEmpty",
			that: PipelineTaskExecutorNameEmpty,
			want: true,
		},
		{
			name: "PipelineTaskExecutorNameSchedulerDefault",
			that: PipelineTaskExecutorNameSchedulerDefault,
			want: true,
		},
		{
			name: "PipelineTaskExecutorNameAPITestDefault",
			that: PipelineTaskExecutorNameAPITestDefault,
			want: true,
		},
		{
			name: "other",
			that: "other",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.that.Check(); got != tt.want {
				t.Errorf("Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipelineTaskExecutorKind_Check(t *testing.T) {
	tests := []struct {
		name string
		that PipelineTaskExecutorKind
		want bool
	}{
		{
			name: "PipelineTaskExecutorKindScheduler",
			that: PipelineTaskExecutorKindScheduler,
			want: true,
		},
		{
			name: "PipelineTaskExecutorKindMemory",
			that: PipelineTaskExecutorKindMemory,
			want: true,
		},
		{
			name: "PipelineTaskExecutorKindAPITest",
			that: PipelineTaskExecutorKindAPITest,
			want: true,
		},
		{
			name: "other",
			that: "other",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.that.Check(); got != tt.want {
				t.Errorf("Check() = %v, want %v", got, tt.want)
			}
		})
	}
}
