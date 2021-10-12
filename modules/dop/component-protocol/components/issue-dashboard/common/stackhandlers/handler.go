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

type StackHandler interface {
	GetStacks() []Stack
	GetIndexer() func(issue interface{}) string

	GetFilterOptions() []filter.PropConditionOption
}

type Stack struct {
	Name  string
	Value string
	Color string
}

var emptyStack = Stack{
	Name:  "empty",
	Value: "",
	Color: "red",
}

func reverseStacks(stacks []Stack) {
	for i, j := 0, len(stacks)-1; i < j; i, j = i+1, j-1 {
		stacks[i], stacks[j] = stacks[j], stacks[i]
	}
}

func getFilterOptions(stacks []Stack, reverse ...bool) []filter.PropConditionOption {
	l := len(stacks)
	options := make([]filter.PropConditionOption, l)
	for i := 0; i < l; i++ {
		j := i
		if len(reverse) > 0 && reverse[0] {
			j = l - 1 - i
		}
		s := stacks[j]
		options[i] = filter.PropConditionOption{
			Label: s.Name,
			Value: s.Value,
		}
	}
	return options
}

type StackRetriever struct {
	reverseStack bool
	issueStateList []dao.IssueState
	issueStageList []apistructs.IssueStage
}

type option func(retriever *StackRetriever)

func NewStackRetriever(options ...option) *StackRetriever {
	retriever := StackRetriever{}
	for _, op := range options {
		op(&retriever)
	}
	return &retriever
}

func WithStacksReversed(reverse bool) option {
	return func(retriever *StackRetriever) {
		retriever.reverseStack = reverse
	}
}

func WithIssueStateList(issueStateList []dao.IssueState) option {
	return func(retriever *StackRetriever) {
		retriever.issueStateList = issueStateList
	}
}

func WithIssueStageList(issueStageList []apistructs.IssueStage) option {
	return func(retriever *StackRetriever) {
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

func (r *StackRetriever) GetRetriever(t string) StackHandler {
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
