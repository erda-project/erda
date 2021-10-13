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

package stackhandlers

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

type SeriesHandler interface {
	GetSeries() []Series
	GetIndexer() func(issue interface{}) string

	GetFilterOptions() []filter.PropConditionOption
}

type Series struct {
	Name  string
	Value string
	Color string
}

var emptySeries = Series{
	Name:  "empty",
	Value: "",
	Color: "red",
}

func reverseSeries(series []Series) {
	for i, j := 0, len(series)-1; i < j; i, j = i+1, j-1 {
		series[i], series[j] = series[j], series[i]
	}
}

func getFilterOptions(series []Series, reverse ...bool) []filter.PropConditionOption {
	l := len(series)
	options := make([]filter.PropConditionOption, l)
	for i := 0; i < l; i++ {
		j := i
		if len(reverse) > 0 && reverse[0] {
			j = l - 1 - i
		}
		s := series[j]
		options[i] = filter.PropConditionOption{
			Label: s.Name,
			Value: s.Value,
		}
	}
	return options
}

type SeriesRetriever struct {
	reverseStack   bool
	issueStateList []dao.IssueState
	issueStageList []apistructs.IssueStage
}

type option func(retriever *SeriesRetriever)

func NewStackRetriever(options ...option) *SeriesRetriever {
	retriever := SeriesRetriever{}
	for _, op := range options {
		op(&retriever)
	}
	return &retriever
}

func WithSeriesReversed(reverse bool) option {
	return func(retriever *SeriesRetriever) {
		retriever.reverseStack = reverse
	}
}

func WithIssueStateList(issueStateList []dao.IssueState) option {
	return func(retriever *SeriesRetriever) {
		retriever.issueStateList = issueStateList
	}
}

func WithIssueStageList(issueStageList []apistructs.IssueStage) option {
	return func(retriever *SeriesRetriever) {
		retriever.issueStageList = issueStageList
	}
}

const (
	Priority   = "Priority"
	Complexity = "Complexity"
	Severity   = "Severity"
	State      = "State"
	Stage      = "Stage"
)

func (r *SeriesRetriever) GetRetriever(t string) SeriesHandler {
	switch t {
	case Priority:
		return NewPriorityStackHandler(r.reverseStack)
	case Complexity:
		return NewComplexityStackHandler(r.reverseStack)
	case Severity:
		return NewSeverityStackHandler(r.reverseStack)
	case State:
		return NewStateStackHandler(r.reverseStack, r.issueStateList)
	case Stage:
		return NewStageStackHandler(r.reverseStack, r.issueStageList)
	default:
		return NewEmptyStackHandler()
	}
}
