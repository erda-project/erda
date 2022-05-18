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
)

func TestMergeRequestInfo_IsJoinTempBranch(t *testing.T) {
	type fields struct {
		JoinTempBranchStatus string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "test success",
			fields: fields{
				JoinTempBranchStatus: JoinTempBranchSuccessStatus,
			},
			want: true,
		},
		{
			name:   "test empty",
			fields: fields{},
			want:   false,
		},
		{
			name: "test other",
			fields: fields{
				JoinTempBranchStatus: "other",
			},
			want: true,
		},
		{
			name: "test failed",
			fields: fields{
				JoinTempBranchStatus: JoinTempBranchFailedStatus,
			},
			want: true,
		},
		{
			name: "test remove",
			fields: fields{
				JoinTempBranchStatus: RemoveFromTempBranchStatus,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			that := MergeRequestInfo{
				JoinTempBranchStatus: tt.fields.JoinTempBranchStatus,
			}
			if got := that.IsJoinTempBranch(); got != tt.want {
				t.Errorf("IsJoinTempBranch() = %v, want %v", got, tt.want)
			}
		})
	}
}
