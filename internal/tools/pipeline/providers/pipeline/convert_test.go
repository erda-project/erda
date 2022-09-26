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

package pipeline

import (
	"testing"
	"time"

	"github.com/alecthomas/assert"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/definition/db"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
)

func Test_ConvertPipeline(t *testing.T) {
	var tables = []struct {
		pipeline spec.Pipeline
	}{
		{
			pipeline: spec.Pipeline{
				PipelineBase: spec.PipelineBase{
					TriggerMode: apistructs.PipelineTriggerModeCron,
				},
			},
		},
		{
			pipeline: spec.Pipeline{
				PipelineBase: spec.PipelineBase{
					TriggerMode: apistructs.PipelineTriggerModeCron,
				},
				PipelineExtra: spec.PipelineExtra{
					Extra: spec.PipelineExtraInfo{
						CronTriggerTime: &[]time.Time{time.Date(2020, 3, 16, 14, 0, 0, 0, time.UTC)}[0],
					},
				},
			},
		},
	}
	var svc = pipelineService{}
	for _, data := range tables {
		dto := svc.ConvertPipeline(&data.pipeline)
		if data.pipeline.Extra.CronTriggerTime != nil {
			assert.Equal(t, dto.TimeCreated.AsTime().Second(), data.pipeline.Extra.CronTriggerTime.Second())
			assert.Equal(t, dto.TimeBegin.AsTime().Second(), data.pipeline.Extra.CronTriggerTime.Second())
		}
	}
}

func Test_transferStatusToAnalyzedFailedIfNeed(t *testing.T) {
	tests := []struct {
		name string
		p    spec.Pipeline
		want apistructs.PipelineStatus
	}{
		{
			name: "analyzed and not abort pipeline",
			p: spec.Pipeline{
				PipelineBase: spec.PipelineBase{
					Status: apistructs.PipelineStatusAnalyzed,
				},
			},
			want: apistructs.PipelineStatusAnalyzed,
		},
		{
			name: "analyzed and abort pipeline",
			p: spec.Pipeline{
				PipelineBase: spec.PipelineBase{
					Status: apistructs.PipelineStatusAnalyzed,
				},
				PipelineExtra: spec.PipelineExtra{
					Extra: spec.PipelineExtraInfo{
						ShowMessage: &basepb.ShowMessage{
							AbortRun: true,
						},
					},
				},
			},
			want: apistructs.PipelineStatusAnalyzeFailed,
		},
		{
			name: "normal pipeline",
			p: spec.Pipeline{
				PipelineBase: spec.PipelineBase{
					Status: apistructs.PipelineStatusRunning,
				},
			},
			want: apistructs.PipelineStatusRunning,
		},
	}
	s := pipelineService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.transferStatusToAnalyzedFailedIfNeed(&tt.p); got != tt.want {
				t.Errorf("transferStatusToAnalyzedFailedIfNeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvert2PagePipeline(t *testing.T) {
	now := time.Now()
	p := &spec.Pipeline{
		PipelineBase: spec.PipelineBase{
			ID:             1,
			PipelineSource: apistructs.PipelineSourceDice,
			TimeBegin:      &now,
			TimeCreated:    &now,
			TimeUpdated:    &now,
			TimeEnd:        &now,
			TriggerMode:    apistructs.PipelineTriggerModeCron,
		},
		PipelineExtra: spec.PipelineExtra{
			CommitDetail: apistructs.CommitDetail{
				CommitID: "1",
			},
			Extra: spec.PipelineExtraInfo{
				CronTriggerTime: &now,
			},
		},
		Labels:     map[string]string{},
		Definition: &db.PipelineDefinition{},
	}
	svc := &pipelineService{}
	pagePipeline := svc.Convert2PagePipeline(p)
	assert.Equal(t, "1", pagePipeline.Commit)
}
