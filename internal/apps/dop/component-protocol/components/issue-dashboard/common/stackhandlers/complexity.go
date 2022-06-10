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
	"context"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/component-protocol/components/issue-dashboard/common/model"
	"github.com/erda-project/erda/internal/apps/dop/dao"
	"github.com/erda-project/erda/internal/tools/openapi/legacy/component-protocol/components/filter"
)

type ComplexityStackHandler struct {
	reverse bool
}

func NewComplexityStackHandler(reverse bool) *ComplexityStackHandler {
	return &ComplexityStackHandler{reverse}
}

var complexityColorMap = map[apistructs.IssueComplexity]string{
	apistructs.IssueComplexityHard:   "red",
	apistructs.IssueComplexityNormal: "yellow",
	apistructs.IssueComplexityEasy:   "green",
}

func (h *ComplexityStackHandler) GetStacks(ctx context.Context) []Stack {
	var stacks []Stack
	for _, i := range []apistructs.IssueComplexity{
		apistructs.IssueComplexityEasy,
		apistructs.IssueComplexityNormal,
		apistructs.IssueComplexityHard,
	} {
		stacks = append(stacks, Stack{
			Name:  cputil.I18n(ctx, string(i)),
			Value: string(i),
			Color: complexityColorMap[i],
		})
	}
	if h.reverse {
		reverseStacks(stacks)
	}
	return stacks
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

func (h *ComplexityStackHandler) GetFilterOptions(ctx context.Context) []filter.PropConditionOption {
	return getFilterOptions(h.GetStacks(ctx), true)
}
