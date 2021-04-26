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

package apistructs

import "testing"

func TestPipelineStatus_AfterPipelineQueue(t *testing.T) {
	tests := []struct {
		name   string
		status PipelineStatus
		want   bool
	}{
		{
			name:   "analyzed",
			status: PipelineStatusAnalyzed,
			want:   false,
		},
		{
			name:   "queue",
			status: PipelineStatusQueue,
			want:   false,
		},
		{
			name:   "running",
			status: PipelineStatusRunning,
			want:   true,
		},
		{
			name:   "success",
			status: PipelineStatusSuccess,
			want:   true,
		},
		{
			name:   "analyzeFailed",
			status: PipelineStatusAnalyzeFailed,
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.AfterPipelineQueue(); got != tt.want {
				t.Errorf("AfterPipelineQueue() = %v, want %v", got, tt.want)
			}
		})
	}
}
