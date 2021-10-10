package stackhandlers

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/model"
	"github.com/erda-project/erda/modules/dop/dao"
)

type StateStackHandler struct {
}

var colorMap = map[apistructs.IssueState][]string {

}

func (h StateStackHandler) GetStacks() []string {
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

func (h StateStackHandler) GetStackColors() []string {
	return []string{"yellow", "blue", "darkseagreen"}
}

func (h StateStackHandler) GetIndexer() func(issue interface{}) string {
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
