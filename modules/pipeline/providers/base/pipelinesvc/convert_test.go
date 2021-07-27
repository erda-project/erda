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

package pipelinesvc

import (
	"testing"
	"time"

	"github.com/bmizerany/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
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
	var svc = PipelineSvc{}
	for _, data := range tables {
		dto := svc.ConvertPipeline(&data.pipeline)
		if data.pipeline.Extra.CronTriggerTime != nil {
			assert.Equal(t, dto.TimeCreated.Second(), data.pipeline.Extra.CronTriggerTime.Second())
			assert.Equal(t, dto.TimeBegin.Second(), data.pipeline.Extra.CronTriggerTime.Second())
		}
	}
}
