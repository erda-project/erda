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

type SeverityStackHandler struct {
}

func (h SeverityStackHandler) GetStacks() []string {
	var stacks []string
	for _, i := range apistructs.IssueSeveritys {
		stacks = append(stacks, i.GetZhName())
	}
	return stacks
}

func (h SeverityStackHandler) GetStackColors() []string {
	return []string{"maroon", "red", "yellow", "darkseagreen", "green"}
}

func (h SeverityStackHandler) GetIndexer() func(issue interface{}) string {
	return func(issue interface{}) string {
		switch issue.(type) {
		case *dao.IssueItem:
			return issue.(*dao.IssueItem).Severity.GetZhName()
		case *model.LabelIssueItem:
			return issue.(*model.LabelIssueItem).Bug.Severity.GetZhName()
		default:
			return ""
		}
	}
}
