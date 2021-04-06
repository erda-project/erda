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

package costtimeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/modules/pipeline/spec"
)

func TestCalculateTaskCostTimeSec(t *testing.T) {
	require.True(t, CalculateTaskCostTimeSec(&spec.PipelineTask{CostTimeSec: 100}) == 100)
	require.True(t, CalculateTaskCostTimeSec(&spec.PipelineTask{CostTimeSec: -1, TimeBegin: time.Time{}}) == -1)
	begin := time.Now()
	time.Sleep(time.Second * 1)
	require.True(t, CalculateTaskCostTimeSec(&spec.PipelineTask{CostTimeSec: -1, TimeBegin: begin}) > 0)
	require.True(t, CalculateTaskCostTimeSec(&spec.PipelineTask{CostTimeSec: -1, TimeBegin: begin, TimeEnd: begin.Add(time.Second * 2)}) == 2)
}

func TestCalculatePipelineCostTimeSec(t *testing.T) {
	require.True(t, CalculatePipelineCostTimeSec(&spec.Pipeline{PipelineBase: spec.PipelineBase{CostTimeSec: 100}}) == 100)
	require.True(t, CalculatePipelineCostTimeSec(&spec.Pipeline{PipelineBase: spec.PipelineBase{CostTimeSec: -1, TimeBegin: &time.Time{}}}) == -1)
	begin := time.Now()
	time.Sleep(time.Second * 1)
	require.True(t, CalculatePipelineCostTimeSec(&spec.Pipeline{PipelineBase: spec.PipelineBase{CostTimeSec: -1, TimeBegin: &begin}}) > 0)
	end := begin.Add(time.Minute * 2)
	require.Equal(t, int64(time.Minute*2/time.Second), CalculatePipelineCostTimeSec(&spec.Pipeline{PipelineBase: spec.PipelineBase{CostTimeSec: -1, TimeBegin: &begin, TimeEnd: &end}}))
}
