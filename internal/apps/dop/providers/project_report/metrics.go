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
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/apistructs"
	iterationdb "github.com/erda-project/erda/internal/apps/dop/dao"
	"github.com/erda-project/erda/internal/core/legacy/model"
)

const (
	metricGroupName = "project_management_report"
)

const (
	labelOrgName     = "org_name"
	labelProjectName = "project_name"

	labelMeta           = "_meta"
	labelMetricScope    = "_metric_scope"
	labelMetricScopeID  = "_metric_scope_id"
	labelIterationID    = "iteration_id"
	labelProjectID      = "project_id"
	labelOrgID          = "org_id"
	labelIterationTitle = "iteration_title"
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

	// task-related metrics
	TaskTotal             uint64
	TaskEstimatedMinute   uint64
	TaskElapsedMinute     uint64
	TaskCompleteSchedule  float64
	TaskAssociatedPercent float64

	// requirement-related metrics
	RequirementTotal             uint64
	RequirementCompleteSchedule  float64
	RequirementAssociatedPercent float64

	// bug-related metrics
	BugTotal               uint64
	SeriousBugPercent      float64
	DemandDesignBugPercent float64
	ReopenBugPercent       float64
	BugCompleteSchedule    float64

	// iteration-related metrics
	IterationAssigneeNum uint64
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
		},
	}
)

var invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// sanitizeLabelName replaces anything that doesn't match
// client_label.LabelNameRE with an underscore.
func sanitizeLabelName(name string) string {
	return invalidNameCharRE.ReplaceAllString(name, "_")
}

func DefaultIterationLabels(iteration *apistructs.Iteration) map[string]string {
	set := map[string]string{
		labelIterationID:    strconv.FormatInt(iteration.ID, 10),
		labelProjectID:      strconv.FormatUint(iteration.ProjectID, 10),
		labelIterationTitle: iteration.Title,
	}
	return set
}