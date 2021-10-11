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
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/model"
	"github.com/erda-project/erda/modules/dop/dao"
)

type ComplexityStackHandler struct {
}

func NewComplexityStackHandler() *ComplexityStackHandler {
	return &ComplexityStackHandler{}
}

func (h *ComplexityStackHandler) GetStacks() []Stack {
	var stacks []Stack
	for _, i := range []apistructs.IssueComplexity{
		apistructs.IssueComplexityHard,
		apistructs.IssueComplexityNormal,
		apistructs.IssueComplexityEasy,
	} {
		stacks = append(stacks, Stack{
			Name:  i.GetZhName(),
			Value: string(i),
		})
	}
	return stacks
}

func (h *ComplexityStackHandler) GetStackColors() []string {
	return []string{"red", "yellow", "green"}
}

func (h *ComplexityStackHandler) GetIndexer() func(issue interface{}) string {
	return func(issue interface{}) string {
		switch issue.(type) {
		case *dao.IssueItem:
			return string(issue.(*dao.IssueItem).Complexity)
		case *model.LabelIssueItem:
			return string(issue.(*model.LabelIssueItem).Bug.Complexity)
		default:
			return ""
		}
	}
}
