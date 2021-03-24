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
	require.True(t, CalculatePipelineCostTimeSec(&spec.Pipeline{CostTimeSec: 100}) == 100)
	require.True(t, CalculatePipelineCostTimeSec(&spec.Pipeline{CostTimeSec: -1, TimeBegin: time.Time{}}) == -1)
	begin := time.Now()
	time.Sleep(time.Second * 1)
	require.True(t, CalculatePipelineCostTimeSec(&spec.Pipeline{CostTimeSec: -1, TimeBegin: begin}) > 0)
	require.Equal(t, int64(time.Minute*2/time.Second), CalculatePipelineCostTimeSec(&spec.Pipeline{CostTimeSec: -1, TimeBegin: begin, TimeEnd: begin.Add(time.Minute * 2)}))
}
