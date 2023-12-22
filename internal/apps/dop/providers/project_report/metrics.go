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

package project_report

import (
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	iterationdb "github.com/erda-project/erda/internal/apps/dop/dao"
	em "github.com/erda-project/erda/internal/apps/dop/providers/efficiency_measure"
	"github.com/erda-project/erda/internal/core/legacy/model"
)

const (
	metricGroupName = "project_management_report"
)

const (
	labelOrgName            = "org_name"
	labelProjectName        = "project_name"
	labelProjectDisplayName = "project_display_name"

	labelMeta               = "_meta"
	labelMetricScope        = "_metric_scope"
	labelMetricScopeID      = "_metric_scope_id"
	labelIterationID        = "iteration_id"
	labelProjectID          = "project_id"
	labelOrgID              = "org_id"
	labelIterationTitle     = "iteration_title"
	labelIterationAssignees = "iteration_assignees"
	labelIterationItemUUID  = "item_ids_uuid"
)

type IterationInfo struct {
	Iteration  *iterationdb.Iteration
	OrgDto     *orgpb.Org
	ProjectDto *model.Project
	Labels     []string

	IterationMetricFields *IterationMetricFields
}

type IterationMetricFields struct {
	IterationID  uint64
	CalculatedAt time.Time
	UUID         string

	// task-related metrics
	TaskTotal                          uint64
	TaskTotalIDs                       []uint64
	TaskEstimatedMinute                uint64
	TaskElapsedMinute                  uint64
	TaskDoneTotal                      uint64
	TaskDoneTotalIDs                   []uint64
	TaskCompleteSchedule               float64
	TaskBeInclusionRequirementTotal    uint64
	TaskBeInclusionRequirementTotalIDs []uint64
	TaskUnAssociatedTotal              uint64
	TaskUnAssociatedTotalIDs           []uint64
	TaskAssociatedPercent              float64
	TaskWorkingTotal                   uint64
	TaskWorkingTotalIDs                []uint64
	TaskEstimatedDayGtOneTotal         uint64
	TaskEstimatedDayGtTwoTotal         uint64
	TaskEstimatedDayGtThreeTotal       uint64

	// requirement-related metrics
	RequirementTotal                  uint64
	RequirementTotalIDs               []uint64
	RequirementDoneTotal              uint64
	RequirementDoneTotalIDs           []uint64
	RequirementCompleteSchedule       float64
	RequirementAssociatedTaskTotal    uint64
	RequirementAssociatedTaskTotalIDs []uint64
	RequirementAssociatedPercent      float64

	// bug-related metrics
	BugTotal                uint64
	BugTotalIDs             []uint64
	SeriousBugTotal         uint64
	SeriousBugTotalIDs      []uint64
	SeriousBugPercent       float64
	DemandDesignBugTotal    uint64
	DemandDesignBugTotalIDs []uint64
	DemandDesignBugPercent  float64
	ReopenBugTotal          uint64
	ReopenBugTotalIDs       []uint64
	ReopenBugPercent        float64
	BugDoneTotal            uint64
	BugDoneTotalIDs         []uint64
	BugUndoneTotal          uint64
	BugUndoneTotalIDs       []uint64
	BugCompleteSchedule     float64
	BugWontfixTotal         uint64
	BugWontfixTotalIDs      []uint64
	BugWithWonfixTotal      uint64
	BugWithWonfixTotalIDs   []uint64

	// iteration-related metrics
	IterationAssigneeNum       uint64
	IterationAssignees         []string
	IterationEstimatedDayTotal float64
	ProjectAssigneeNum         uint64
}

// IsValid returns true if the IterationMetricFields is valid.
// we need to ensure that IterationMetricFields is the data of the day to avoid double calculation
func (i *IterationMetricFields) IsValid() bool {
	return i.CalculatedAt.Day() == time.Now().Day()
}

var (
	allIterationMetrics = []iterationMetric{
		{
			name:      "iteration_task_total",
			help:      "Cumulative task type issue iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = float64(iterationInfo.IterationMetricFields.TaskTotal)
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.TaskTotalIDs)
			},
		},
		{
			name:      "iteration_requirement_total",
			help:      "Cumulative requirement type issue iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = float64(iterationInfo.IterationMetricFields.RequirementTotal)
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.RequirementDoneTotalIDs)
			},
		},
		{
			name:      "iteration_bug_total",
			help:      "Cumulative bug type issue iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = float64(iterationInfo.IterationMetricFields.BugTotal)
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.BugTotalIDs)
			},
		},
		{
			name:      "iteration_task_estimated_minute",
			help:      "Cumulative estimated minute of task type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value uint64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.TaskEstimatedMinute
				}
				return metricValues{
					{
						value:     float64(value),
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_task_elapsed_minute",
			help:      "Cumulative elapsed minute of task type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value uint64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.TaskElapsedMinute
				}
				return metricValues{
					{
						value:     float64(value),
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_task_unassociated_total",
			help:      "Cumulative unassociated with requirement task type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = float64(iterationInfo.IterationMetricFields.TaskUnAssociatedTotal)
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.TaskUnAssociatedTotalIDs)
			},
		},
		{
			name:      "iteration_task_complete_schedule",
			help:      "Cumulative complete schedule of task type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.TaskCompleteSchedule
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_task_working_total",
			help:      "Cumulative working task type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = float64(iterationInfo.IterationMetricFields.TaskWorkingTotal)
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.TaskWorkingTotalIDs)
			},
		},
		{
			name:      "iteration_task_estimated_day_gt_one_total",
			help:      "Cumulative estimated day greater than one of task type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = float64(iterationInfo.IterationMetricFields.TaskEstimatedDayGtOneTotal)
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_task_estimated_day_gt_two_total",
			help:      "Cumulative estimated day greater than two of task type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = float64(iterationInfo.IterationMetricFields.TaskEstimatedDayGtTwoTotal)
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_task_estimated_day_gt_three_total",
			help:      "Cumulative estimated day greater than three of task type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = float64(iterationInfo.IterationMetricFields.TaskEstimatedDayGtThreeTotal)
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_task_associated_percent",
			help:      "Accumulate the proportion of the number of requirements associated with the current task",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.TaskAssociatedPercent
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_requirement_complete_schedule",
			help:      "Cumulative complete schedule of requirement type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.RequirementCompleteSchedule
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_requirement_associated_percent",
			help:      "Accumulate the proportion of the number of requirements associated with the task",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.RequirementAssociatedPercent
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_serious_bug_percent",
			help:      "Cumulative fatal/serious bug percent of bug type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.SeriousBugPercent
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_demand_design_bug_percent",
			help:      "Cumulative demand/design bug percent of bug type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.DemandDesignBugPercent
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_reopen_bug_percent",
			help:      "Cumulative reopen bug percent of bug type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.ReopenBugPercent
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_bug_complete_schedule",
			help:      "Cumulative complete schedule of bug type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.BugCompleteSchedule
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_bug_wontfix_total",
			help:      "Cumulative wontfix bug total of bug type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = float64(iterationInfo.IterationMetricFields.BugWontfixTotal)
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.BugWontfixTotalIDs)
			},
		},
		{
			name:      "iteration_bug_with_wonfix_total",
			help:      "Cumulative with_wonfix bug total of bug type issue in iteration.",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = float64(iterationInfo.IterationMetricFields.BugWithWonfixTotal)
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.BugWithWonfixTotalIDs)
			},
		},
		{
			name:      "iteration_assignee_total",
			help:      "The number of people who still have unfinished tasks in the project as of now",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value uint64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.IterationAssigneeNum
				}
				return metricValues{
					{
						value:     float64(value),
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "project_assignee_total",
			help:      "The number of people who still have unfinished tasks in the project as of now",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value uint64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.ProjectAssigneeNum
				}
				return metricValues{
					{
						value:     float64(value),
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
		{
			name:      "iteration_task_done_total",
			help:      "The number of tasks that have been completed in the iteration",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value uint64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.TaskDoneTotal
				}
				return metricValues{
					{
						value:     float64(value),
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.TaskDoneTotalIDs)
			},
		},
		{
			name:      "iteration_task_inclusion_requirement_total",
			help:      "The number of tasks that are associated by the requirement in the iteration",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value uint64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.TaskBeInclusionRequirementTotal
				}
				return metricValues{
					{
						value:     float64(value),
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.TaskBeInclusionRequirementTotalIDs)
			},
		},
		{
			name:      "iteration_requirement_done_total",
			help:      "The number of requirements that have been completed in the iteration",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value uint64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.RequirementDoneTotal
				}
				return metricValues{
					{
						value:     float64(value),
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.RequirementDoneTotalIDs)
			},
		},
		{
			name:      "iteration_requirement_associated_task_total",
			help:      "The number of requirement have associated tasks in the iteration",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value uint64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.RequirementAssociatedTaskTotal
				}
				return metricValues{
					{
						value:     float64(value),
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.RequirementAssociatedTaskTotalIDs)
			},
		},
		{
			name:      "iteration_serious_bug_total",
			help:      "The number of fatal/serious bug in the iteration",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value uint64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.SeriousBugTotal
				}
				return metricValues{
					{
						value:     float64(value),
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.SeriousBugTotalIDs)
			},
		},
		{
			name:      "iteration_bug_undone_total",
			help:      "The number of bug that have not been completed in the iteration",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = float64(iterationInfo.IterationMetricFields.BugUndoneTotal)
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.BugUndoneTotalIDs)
			},
		},
		{
			name:      "iteration_demand_design_bug_total",
			help:      "The number of demand design bug in the iteration",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value uint64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.DemandDesignBugTotal
				}
				return metricValues{
					{
						value:     float64(value),
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.DemandDesignBugTotalIDs)
			},
		},
		{
			name:      "iteration_reopen_bug_total",
			help:      "The number of reopen bug in the iteration",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value uint64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.ReopenBugTotal
				}
				return metricValues{
					{
						value:     float64(value),
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.ReopenBugTotalIDs)
			},
		},
		{
			name:      "iteration_bug_done_total",
			help:      "The number of bug that have been completed in the iteration",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value uint64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.BugDoneTotal
				}
				return metricValues{
					{
						value:     float64(value),
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return em.GetMetricsItemIDs(i.IterationMetricFields.BugDoneTotalIDs)
			},
		},
		{
			name:      "iteration_estimated_day_total",
			help:      "The total number estimated days of the iteration",
			valueType: prometheus.CounterValue,
			getValues: func(iterationInfo *IterationInfo) metricValues {
				var value float64
				if iterationInfo.IterationMetricFields != nil {
					value = iterationInfo.IterationMetricFields.IterationEstimatedDayTotal
				}
				return metricValues{
					{
						value:     value,
						timestamp: time.Now(),
					},
				}
			},
			getMetricsItemIDs: func(i *IterationInfo) string {
				return "0"
			},
		},
	}
)

var invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// sanitizeLabelName replaces anything that doesn't match
// client_label.LabelNameRE with an underscore.
func sanitizeLabelName(name string) string {
	return invalidNameCharRE.ReplaceAllString(name, "_")
}
