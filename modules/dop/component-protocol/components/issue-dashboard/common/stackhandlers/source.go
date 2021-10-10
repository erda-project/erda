package stackhandlers

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/model"
	"github.com/erda-project/erda/modules/dop/dao"
)

type SourceStackHandler struct {
}

func (h SourceStackHandler) GetStacks() []string {
	var stacks []string
	for _, i := range []apistructs.IssueComplexity{
		apistructs.IssueComplexityHard,
		apistructs.IssueComplexityNormal,
		apistructs.IssueComplexityEasy,
	} {
		stacks = append(stacks, i.GetZhName())
	}
	return stacks
}

func (h SourceStackHandler) GetStackColors() []string {
	return []string{"red", "yellow", "green"}
}

func (h SourceStackHandler) GetIndexer() func(issue interface{}) string {
	return func(issue interface{}) string {
		switch issue.(type) {
		case *dao.IssueItem:
			return issue.(*dao.IssueItem).Complexity.GetZhName()
		case *model.LabelIssueItem:
			return issue.(*model.LabelIssueItem).Bug.Complexity.GetZhName()
		default:
			return ""
		}
	}
}
