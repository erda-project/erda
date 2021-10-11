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
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/model"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

type StageStackHandler struct {
	issueStageList []dao.IssueStage
}

func NewStageStackHandler(issueStageList []dao.IssueStage) *StageStackHandler {
	return &StageStackHandler{
		issueStageList: issueStageList,
	}
}

var stageColorList = []string{"red", "yellow", "green"}

func (h *StageStackHandler) GetStacks() []Stack {
	l := len(stageColorList)
	var stacks []Stack
	for idx, i := range h.issueStageList {
		stacks = append(stacks, Stack{
			Name:  i.Name,
			Value: i.Value,
			Color: stageColorList[idx%l],
		})
	}
	if len(stacks) == 0 {
		stacks = append(stacks, emptyStack)
	}
	return stacks
}

func (h *StageStackHandler) GetIndexer() func(issue interface{}) string {
	return func(issue interface{}) string {
		switch issue.(type) {
		case *dao.IssueItem:
			return issue.(*dao.IssueItem).Stage
		case *model.LabelIssueItem:
			return issue.(*model.LabelIssueItem).Bug.Stage
		default:
			return ""
		}
	}
}

func (h *StageStackHandler) GetFilterOptions() []filter.PropConditionOption {
	return getFilterOptions(h.GetStacks())
}
