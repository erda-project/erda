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

package pipelineTable

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/providers/component-protocol/components/commodel"
	"github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/stretchr/testify/assert"
)

func TestMakeInode(t *testing.T) {
	tt := []struct {
		appName      string
		definition   *pb.PipelineDefinition
		appNameIDMap *apistructs.GetAppIDByNamesResponseData
		want         string
	}{
		{
			appName: "erda",
			definition: &pb.PipelineDefinition{
				Ref:      "master",
				Path:     "",
				FileName: "pipeline.yml",
			},
			appNameIDMap: &apistructs.GetAppIDByNamesResponseData{AppNameToID: map[string]int64{
				"erda": 1,
			}},
			want: "1/1/tree/master/pipeline.yml",
		},
		{
			appName: "erda",
			definition: &pb.PipelineDefinition{
				Ref:      "master",
				Path:     ".erda/pipelines",
				FileName: "pipeline.yml",
			},
			appNameIDMap: &apistructs.GetAppIDByNamesResponseData{AppNameToID: map[string]int64{
				"erda": 1,
			}},
			want: "1/1/tree/master/.erda/pipelines/pipeline.yml",
		},
		{
			appName: "",
			definition: &pb.PipelineDefinition{
				Ref:      "master",
				Path:     ".erda/pipelines",
				FileName: "pipeline.yml",
			},
			appNameIDMap: &apistructs.GetAppIDByNamesResponseData{AppNameToID: map[string]int64{
				"erda": 1,
			}},
			want: "",
		},
		{
			appName: "erda",
			definition: &pb.PipelineDefinition{
				Ref:      "master",
				Path:     ".erda/pipelines",
				FileName: "pipeline.yml",
			},
			appNameIDMap: nil,
			want:         "",
		},
	}
	p := &PipelineTable{
		InParams: &InParams{
			ProjectID: 1,
		},
	}
	for _, v := range tt {
		assert.Equal(t, v.want, p.makeInode(v.appName, v.definition, v.appNameIDMap))
	}
}

func Test_getStatus(t *testing.T) {
	tests := []struct {
		name   string
		status apistructs.PipelineStatus
		want   commodel.UnifiedStatus
	}{
		// TODO: Add test cases.
		{
			name:   "analyzed",
			status: apistructs.PipelineStatusAnalyzed,
			want:   commodel.DefaultStatus,
		},
		{
			name:   "running",
			status: apistructs.PipelineStatusRunning,
			want:   commodel.ProcessingStatus,
		},
		{
			name:   "canceling",
			status: apistructs.PipelineStatusCanceling,
			want:   commodel.ProcessingStatus,
		},
		{
			name:   "failed",
			status: apistructs.PipelineStatusFailed,
			want:   commodel.ErrorStatus,
		},
		{
			name:   "AnalyzeFailed",
			status: apistructs.PipelineStatusAnalyzeFailed,
			want:   commodel.ErrorStatus,
		},
		{
			name:   "AnalyzeFailed",
			status: apistructs.PipelineStatusAnalyzeFailed,
			want:   commodel.ErrorStatus,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStatus(tt.status); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
