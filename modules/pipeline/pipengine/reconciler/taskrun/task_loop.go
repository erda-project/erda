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

package taskrun

import (
	"fmt"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pexpr"
	"github.com/erda-project/erda/modules/pipeline/pexpr/pexpr_params"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/loop"
)

// handleTaskLoop Determine whether the task needs to be looped; if necessary, adjust the task state and wait for the thinking time
func (tr *TaskRun) handleTaskLoop() error {
	// not end state, skip
	if !tr.Task.Status.IsEndStatus() {
		return nil
	}

	// The end state of the pipeline does not loop tasks
	tr.EnsureFetchLatestPipelineStatus()
	if tr.QueriedPipelineStatus.IsEndStatus() {
		rlog.TWarnf(tr.P.ID, tr.Task.ID, "pipeline is already end status (%s), not try to loop task", tr.QueriedPipelineStatus)
		return nil
	}

	// No loop configuration skip
	if tr.Task.Extra.LoopOptions == nil || tr.Task.Extra.LoopOptions.CalculatedLoop == nil {
		return nil
	}
	loopOpt := tr.Task.Extra.LoopOptions.CalculatedLoop
	loopedTimes := tr.Task.Extra.LoopOptions.LoopedTimes
	expr := loopOpt.Break
	// no Strategy use Default strategy
	if loopOpt.Strategy == nil {
		loopOpt.Strategy = &apistructs.PipelineTaskDefaultLoopStrategy
	}
	// Determine whether the exit conditions are still not met
	params := pexpr_params.GenerateParamsFromTask(tr.P.ID, tr.Task.ID, tr.Task.Status)
	result, err := pexpr.Eval(expr, params)
	if err != nil {
		return fmt.Errorf("loop break expr %s evaluate failed, err: %v", expr, err)
	}
	// Meet the exit conditions, exit the loop
	if t, ok := result.(bool); ok && t {
		rlog.TWarnf(tr.P.ID, tr.Task.ID, "loop break expr %s evaluate result is true, break loop", expr)
		return nil
	}
	// The maximum number of cycles has been reached, exit the loop
	if loopOpt.Strategy.MaxTimes != -1 && int64(loopedTimes) >= loopOpt.Strategy.MaxTimes {
		rlog.TDebugf(tr.P.ID, tr.Task.ID, "loop reached max times %d, stop loop", loopedTimes)
		return nil
	}
	rlog.TDebugf(tr.P.ID, tr.Task.ID, "loop break expr %s evaluate result is false, continue loop", expr)

	// reportTaskForLoop report task before resetTaskForLoop to avoid missing task info
	if err := tr.reportTaskForLoop(); err != nil {
		rlog.Errorf("failed to report task-loop, pipelineID: %d, taskID: %d, err: %v", tr.P.ID, tr.Task.ID, err)
	}

	tr.resetTaskForLoop()
	return nil
}

// reportTaskForLoop record looped task info
func (tr *TaskRun) reportTaskForLoop() error {
	if tr.Task.Extra.LoopOptions == nil {
		return nil
	}
	meta := map[string]interface{}{
		fmt.Sprintf("task-%d-loop-%d", tr.Task.ID, tr.Task.Extra.LoopOptions.LoopedTimes): *tr.Task,
	}
	return tr.DBClient.CreatePipelineReport(&spec.PipelineReport{
		PipelineID: tr.P.ID,
		Type:       apistructs.PipelineReportLoopMetaKey,
		Meta:       meta,
	})
}

func (tr *TaskRun) resetTaskForLoop() {
	// Calculate sleep time
	strategy := tr.Task.Extra.LoopOptions.CalculatedLoop.Strategy
	interval := loop.New(
		loop.WithInterval(time.Second*time.Duration(strategy.IntervalSec)),
		loop.WithDeclineRatio(strategy.DeclineRatio),
		loop.WithDeclineLimit(time.Second*time.Duration(strategy.DeclineLimitSec)),
	).CalculateInterval(tr.Task.Extra.LoopOptions.LoopedTimes)
	rlog.TDebugf(tr.P.ID, tr.Task.ID, "sleep %s before loop", interval.String())
	time.Sleep(interval)

	// sleep time may be very long, after waiting, check the latest status again
	tr.EnsureFetchLatestPipelineStatus()
	if tr.QueriedPipelineStatus.IsEndStatus() {
		rlog.TWarnf(tr.P.ID, tr.Task.ID, "pipeline is already end status (%s), not loop task after sleep", tr.QueriedPipelineStatus)
		return
	}

	// reset task status
	tr.Task.Extra.LoopOptions.LoopedTimes++
	tr.Task.Status = apistructs.PipelineStatusAnalyzed
	// reset time for loop, all based on the last time
	tr.Task.CostTimeSec = -1
	tr.Task.QueueTimeSec = -1
	tr.Task.Extra.TimeBeginQueue = time.Time{}
	tr.Task.Extra.TimeEndQueue = time.Time{}
	tr.Task.TimeBegin = time.Time{}
	tr.Task.TimeEnd = time.Time{}
	// reset task result
	tr.Task.Result = apistructs.PipelineTaskResult{}
	// reset volume
	tr.Task.Context = spec.PipelineTaskContext{}
	tr.Task.Extra.Volumes = nil
	// reset tr flag
	tr.FakeTimeout = false
	tr.QuitQueueTimeout = false
	tr.QuitWaitTimeout = false
	tr.StopQueueLoop = false
	tr.StopWaitLoop = false
}
