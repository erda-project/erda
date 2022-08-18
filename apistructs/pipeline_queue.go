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

type EnqueueConditionType string

const (
	EnqueueConditionSkipAlreadyRunningLimit EnqueueConditionType = "skip_already_running_limit"
)

func (e EnqueueConditionType) String() string {
	return string(e)
}

func (e EnqueueConditionType) IsSkipAlreadyRunningLimit() bool {
	return e == EnqueueConditionSkipAlreadyRunningLimit
}

const PipelinePreCheckResultContextKey = "precheck_result"

type PipelineQueueMode string

var (
	PipelineQueueModeStrict PipelineQueueMode = "STRICT"
	PipelineQueueModeLoose  PipelineQueueMode = "LOOSE"
)

func (m PipelineQueueMode) String() string { return string(m) }
func (m PipelineQueueMode) IsValid() bool {
	switch m {
	case PipelineQueueModeStrict, PipelineQueueModeLoose:
		return true
	default:
		return false
	}
}

// ScheduleStrategyInsidePipelineQueue represents the schedule strategy of workflows inside a queue.
type ScheduleStrategyInsidePipelineQueue string

var (
	ScheduleStrategyInsidePipelineQueueOfFIFO ScheduleStrategyInsidePipelineQueue = "FIFO"
)

func (strategy ScheduleStrategyInsidePipelineQueue) String() string {
	return string(strategy)
}

func (strategy ScheduleStrategyInsidePipelineQueue) IsValid() bool {
	switch strategy {
	case ScheduleStrategyInsidePipelineQueueOfFIFO:
		return true
	default:
		return false
	}
}

var (
	PipelineQueueDefaultPriority         int64 = 10
	PipelineQueueDefaultScheduleStrategy       = ScheduleStrategyInsidePipelineQueueOfFIFO
	PipelineQueueDefaultMode                   = PipelineQueueModeLoose
	PipelineQueueDefaultConcurrency      int64 = 1
)

// PipelineQueueValidateResult represents queue validate result.
type PipelineQueueValidateResult struct {
	Success     bool                      `json:"success"`
	Reason      string                    `json:"reason"`
	IsEnd       bool                      `json:"isEnd"`
	RetryOption *QueueValidateRetryOption `json:"retryOption"`
}

type QueueValidateRetryOption struct {
	IntervalSecond      uint64 `json:"intervalSecond"`
	IntervalMillisecond uint64 `json:"intervalMillisecond"`
}

func (r PipelineQueueValidateResult) IsSuccess() bool {
	return r.Success
}
func (r PipelineQueueValidateResult) IsFailed() bool {
	return !r.Success
}

func (r PipelineQueueValidateResult) IsEndStatus() bool {
	return r.IsEnd
}
