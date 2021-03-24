package reconciler

import (
	"fmt"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pexpr"
	"github.com/erda-project/erda/modules/pipeline/pexpr/pexpr_params"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/taskrun"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/loop"
)

// handleTaskLoop 判断 task 是否需要循环；若需要，则调整 task 状态，并等待思考时间
func handleTaskLoop(tr *taskrun.TaskRun) error {
	// pipeline 终态则不循环 task
	tr.EnsureFetchLatestPipelineStatus()
	if tr.QueriedPipelineStatus.IsEndStatus() {
		rlog.TWarnf(tr.P.ID, tr.Task.ID, "pipeline is already end status (%s), not try to loop task", tr.QueriedPipelineStatus)
		return nil
	}

	// 无循环配置
	if tr.Task.Extra.LoopOptions == nil || tr.Task.Extra.LoopOptions.CalculatedLoop == nil {
		return nil
	}
	loopOpt := tr.Task.Extra.LoopOptions.CalculatedLoop
	loopedTimes := tr.Task.Extra.LoopOptions.LoopedTimes
	expr := loopOpt.Break
	// 默认策略
	if loopOpt.Strategy == nil {
		loopOpt.Strategy = &apistructs.PipelineTaskDefaultLoopStrategy
	}
	// 判断是否仍不满足退出条件
	params := pexpr_params.GenerateParamsFromTask(tr.P.ID, tr.Task.ID)
	result, err := pexpr.Eval(expr, params)
	if err != nil {
		return fmt.Errorf("loop break expr %s evaluate failed, err: %v", expr, err)
	}
	// 满足退出条件，退出循环
	if t, ok := result.(bool); ok && t {
		rlog.TWarnf(tr.P.ID, tr.Task.ID, "loop break expr %s evaluate result is true, break loop", expr)
		return nil
	}
	// 已达最大循环次数，退出循环
	if loopOpt.Strategy.MaxTimes != -1 && int64(loopedTimes) >= loopOpt.Strategy.MaxTimes {
		rlog.TDebugf(tr.P.ID, tr.Task.ID, "loop reached max times %d, stop loop", loopedTimes)
		return nil
	}
	// 继续循环
	rlog.TDebugf(tr.P.ID, tr.Task.ID, "loop break expr %s evaluate result is false, continue loop", expr)

	resetTaskForLoop(tr)

	return nil
}

func resetTaskForLoop(tr *taskrun.TaskRun) {
	// 计算思考时间
	strategy := tr.Task.Extra.LoopOptions.CalculatedLoop.Strategy
	interval := loop.New(
		loop.WithInterval(time.Second*time.Duration(strategy.IntervalSec)),
		loop.WithDeclineRatio(strategy.DeclineRatio),
		loop.WithDeclineLimit(time.Second*time.Duration(strategy.DeclineLimitSec)),
	).CalculateInterval(tr.Task.Extra.LoopOptions.LoopedTimes)
	// 等待思考时间
	rlog.TDebugf(tr.P.ID, tr.Task.ID, "sleep %s before loop", interval.String())
	time.Sleep(interval)

	// 思考时间可能很长，等待结束后再次校验最新状态
	// pipeline 终态则不循环 task
	tr.EnsureFetchLatestPipelineStatus()
	if tr.QueriedPipelineStatus.IsEndStatus() {
		rlog.TWarnf(tr.P.ID, tr.Task.ID, "pipeline is already end status (%s), not loop task after sleep", tr.QueriedPipelineStatus)
		return
	}

	// 重置任务状态，重新开始执行
	tr.Task.Extra.LoopOptions.LoopedTimes++
	tr.Task.Status = apistructs.PipelineStatusAnalyzed
	// 重置时间 for loop，全部以最后一次时间为准
	tr.Task.CostTimeSec = -1
	tr.Task.QueueTimeSec = -1
	tr.Task.Extra.TimeBeginQueue = time.Time{}
	tr.Task.Extra.TimeEndQueue = time.Time{}
	tr.Task.TimeBegin = time.Time{}
	tr.Task.TimeEnd = time.Time{}
	// 重置任务结果
	tr.Task.Result = apistructs.PipelineTaskResult{}
	// 重置 Volume
	tr.Task.Context = spec.PipelineTaskContext{}
	tr.Task.Extra.Volumes = nil
	// 更新
	tr.Update()
}
