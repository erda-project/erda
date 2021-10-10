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

package common

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

type FrontendConditions struct {
	IterationIDs []int64  `json:"iteration,omitempty"`
	AssigneeIDs  []string `json:"member,omitempty"`
}

type FilterConditions struct {
	Type  string   `json:"type,omitempty"`
	Value []string `json:"value,omitempty"`
}

const (
	Priority   = "Priority"
	Complexity = "Complexity"
	Severity   = "Severity"
)

var ConditionMap = map[string][]filter.PropConditionOption{
	Priority: {
		{
			Label: "紧急",
			Value: apistructs.IssuePriorityUrgent,
		},
		{
			Label: "高",
			Value: apistructs.IssuePriorityHigh,
		},
		{
			Label: "中",
			Value: apistructs.IssuePriorityNormal,
		},
		{
			Label: "低",
			Value: apistructs.IssuePriorityLow,
		},
	},
	Complexity: {
		{
			Label: "复杂",
			Value: apistructs.IssueComplexityHard,
		},
		{
			Label: "中",
			Value: apistructs.IssueComplexityNormal,
		},
		{
			Label: "容易",
			Value: apistructs.IssueComplexityEasy,
		},
	},
	Severity: {
		{
			Label: "致命",
			Value: apistructs.IssueSeverityFatal,
		},
		{
			Label: "严重",
			Value: apistructs.IssueSeveritySerious,
		},
		{
			Label: "一般",
			Value: apistructs.IssueSeverityNormal,
		},
		{
			Label: "轻微",
			Value: apistructs.IssueSeveritySlight,
		},
		{
			Label: "建议",
			Value: apistructs.IssueSeverityLow,
		},
	},
}
