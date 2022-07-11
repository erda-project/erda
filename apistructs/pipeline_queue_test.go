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
	"testing"

	"github.com/bmizerany/assert"
)

func TestScheduleStrategyInsidePipelineQueue_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		strategy ScheduleStrategyInsidePipelineQueue
		want     bool
	}{
		{
			name:     "invalid empty strategy",
			strategy: "",
			want:     false,
		},
		{
			name:     "invalid unknown strategy",
			strategy: "xxx",
			want:     false,
		},
		{
			name:     "valid fifo strategy",
			strategy: ScheduleStrategyInsidePipelineQueueOfFIFO,
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.strategy.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScheduleStrategyInsidePipelineQueue_String(t *testing.T) {
	tests := []struct {
		name     string
		strategy ScheduleStrategyInsidePipelineQueue
		want     string
	}{
		{
			name:     "fifo",
			strategy: ScheduleStrategyInsidePipelineQueueOfFIFO,
			want:     "FIFO",
		},
		{
			name:     "unknown",
			strategy: "xxx",
			want:     "xxx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.strategy.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsEndStatus(t *testing.T) {
	result := PipelineQueueValidateResult{
		IsEnd: true,
	}
	assert.Equal(t, true, result.IsEndStatus())
	result.IsEnd = false
	assert.Equal(t, false, result.IsEndStatus())
}
