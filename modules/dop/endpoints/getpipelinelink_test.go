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

package endpoints

import (
	"fmt"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

func TestGetPipelineLink(t *testing.T) {
	type args struct {
		p      apistructs.PipelineDTO
		ctxMap map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test_empty",
			args: args{
				p: apistructs.PipelineDTO{
					OrgName:       "terminus",
					ProjectID:     123,
					ApplicationID: 123123,
				},
				ctxMap: map[string]interface{}{
					apierrors.ErrParallelRunPipeline.Error(): fmt.Sprintf("%d", 777777),
				},
			},
			want:    "/terminus/dop/projects/123/apps/123123/pipeline?pipelineID=777777",
			wantErr: false,
		},
		{
			name: "test_empty",
			args: args{
				p: apistructs.PipelineDTO{
					OrgName:       "t2erminus",
					ProjectID:     123,
					ApplicationID: 123123,
				},
				ctxMap: map[string]interface{}{
					apierrors.ErrParallelRunPipeline.Error(): fmt.Sprintf("%d", 66666),
				},
			},
			want:    "/t2erminus/dop/projects/123/apps/123123/pipeline?pipelineID=66666",
			wantErr: false,
		},
		{
			name: "test_empty",
			args: args{
				p: apistructs.PipelineDTO{
					OrgName:       "terminus",
					ProjectID:     13,
					ApplicationID: 123123,
				},
				ctxMap: map[string]interface{}{
					apierrors.ErrParallelRunPipeline.Error(): fmt.Sprintf("%d", 777777),
				},
			},
			want:    "/terminus/dop/projects/13/apps/123123/pipeline?pipelineID=777777",
			wantErr: false,
		},
		{
			name: "test_empty",
			args: args{
				p: apistructs.PipelineDTO{
					OrgName:       "terminus",
					ProjectID:     1234,
					ApplicationID: 123123,
				},
				ctxMap: map[string]interface{}{
					apierrors.ErrParallelRunPipeline.Error(): fmt.Sprintf("%d", 77777),
				},
			},
			want:    "/terminus/dop/projects/1234/apps/123123/pipeline?pipelineID=77777",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetPipelineLink(tt.args.p, tt.args.ctxMap)
			if got != tt.want {
				t.Errorf("GetPipelineLink() got = %v, want %v", got, tt.want)
			}
			if !got1 {
				t.Errorf("GetPipelineLink() got1 = %v, want %v", got1, true)
			}
		})
	}
}
