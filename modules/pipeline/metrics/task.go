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

package metrics

// import (
// 	"strconv"

// 	"github.com/erda-project/erda/modules/pipeline/spec"
// 	"terminus.io/dice/telemetry/metrics"
// 	"terminus.io/dice/telemetry/promxp"
// )

// const (
// 	fieldTaskTotal      = "task_total"
// 	fieldTaskProcessing = "task_processing"

// 	labelTaskID     = "task_id"
// 	labelActionType = "action_type"

// 	labelTaskStatus = "task_status"
// )

// var (
// 	taskCounterLabels    = []string{"task_status", "execute_cluster", "action_type"}
// 	taskProcessingLabels = []string{"execute_cluster", "action_type"}
// )

// // TaskCounterTotalAdd 某时间段内累计执行次数、成功次数、失败次数
// var TaskCounterTotalAdd = func(task spec.PipelineTask, value float64) {
// 	if disableMetrics {
// 		return
// 	}
// 	defer func() {
// 		if r := recover(); r != nil {
// 			taskErrorLog(task, "[alert] failed to do metric TaskCounterTotalAdd, err: %v", r)
// 		}
// 	}()
// 	labelValues := []string{task.Status.String(), task.Extra.ClusterName, task.Type}
// 	taskCounterTotal.WithLabelValues(labelValues...).Add(value)
// 	taskDebugLog(task, "metric: TaskCounterTotalAdd, value: %v, labelValues: %v", value, labelValues)
// }

// var taskCounterTotal = promxp.RegisterAutoResetCounterVec(
// 	fieldTaskTotal,
// 	"task total counter",
// 	map[string]string{},
// 	taskCounterLabels,
// )

// // TaskGaugeProcessingAdd 正在处理中的个数
// var TaskGaugeProcessingAdd = func(task spec.PipelineTask, value float64) {
// 	if disableMetrics {
// 		return
// 	}
// 	defer func() {
// 		if r := recover(); r != nil {
// 			taskErrorLog(task, "[alert] failed to do metric TaskGaugeProcessingAdd, err: %v", r)
// 		}
// 	}()
// 	labelValues := []string{task.Extra.ClusterName, task.Type}
// 	taskGaugeProcessing.WithLabelValues(labelValues...).Add(value)
// 	taskDebugLog(task, "metric: TaskGaugeProcessingAdd, value: %v, labelValues: %v", value, labelValues)
// }

// var taskGaugeProcessing = promxp.RegisterGaugeVec(
// 	fieldTaskProcessing,
// 	"processing task",
// 	nil,
// 	taskProcessingLabels,
// )

// // TaskEndEvent 终态时推送事件
// var TaskEndEvent = func(task spec.PipelineTask, p *spec.Pipeline) {
// 	if disableMetrics {
// 		return
// 	}
// 	defer func() {
// 		if r := recover(); r != nil {
// 			taskErrorLog(task, "[alert] failed to do event TaskEndEvent, err: %v", r)
// 		}
// 	}()
// 	request := metrics.CreateBulkMetricRequest()
// 	request.Add("dice_pipeline_task", generateActionEventTags(task, p), generateActionEventFields(task))
// 	if err := bulkClient.Push(request); err != nil {
// 		taskErrorLog(task, "[alert] failed to push task bulk event, err: %v", err)
// 	}
// 	taskDebugLog(task, "send task end event success")
// }

// func generateActionEventTags(task spec.PipelineTask, p *spec.Pipeline) map[string]string {
// 	tags := map[string]string{
// 		labelPipelineID:     strconv.FormatUint(task.PipelineID, 10),
// 		labelTaskID:         strconv.FormatUint(task.ID, 10),
// 		labelActionType:     task.Type,
// 		labelExecuteCluster: task.Extra.ClusterName,
// 		labelTaskStatus:     task.Status.String(),

// 		labelPipelineSource:  labelPipelineSource,
// 		labelPipelineYmlName: p.PipelineYmlName,
// 	}
// 	envs := task.Extra.PrivateEnvs
// 	if envs == nil {
// 		return tags
// 	}

// 	for k, v := range envs {
// 		tags[k] = v
// 	}
// 	return tags
// }

// func generateActionEventFields(task spec.PipelineTask) map[string]interface{} {
// 	return map[string]interface{}{
// 		num: 1,
// 	}
// }
