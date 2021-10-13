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
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

type StageStackHandler struct {
	reverse        bool
	issueStageList []apistructs.IssueStage
}

func NewStageStackHandler(reverse bool, issueStageList []apistructs.IssueStage) *StageStackHandler {
	return &StageStackHandler{
		reverse:        reverse,
		issueStageList: issueStageList,
	}
}

var stageColorList = []string{"red", "yellow", "green"}

func (h *StageStackHandler) GetSeries() []Series {
	l := len(stageColorList)
	var stacks []Series
	for idx, i := range h.issueStageList {
		stacks = append(stacks, Series{
			Name:  i.Name,
			Value: i.Value,
			Color: stageColorList[idx%l],
		})
	}
	if h.reverse {
		reverseSeries(stacks)
	}
	if len(stacks) == 0 {
		stacks = append(stacks, emptySeries)
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
	return getFilterOptions(h.GetSeries())
}
