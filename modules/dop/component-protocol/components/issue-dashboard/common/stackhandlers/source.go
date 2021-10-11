package stackhandlers

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/model"
	"github.com/erda-project/erda/modules/dop/dao"
)

type SourceStackHandler struct {
}

func NewSourceStackHandler() *SourceStackHandler {
	return &SourceStackHandler{}
}

func (h *SourceStackHandler) GetStacks() []Stack {
	var stacks []Stack
	for _, i := range []apistructs.IssueComplexity{
		apistructs.IssueComplexityHard,
		apistructs.IssueComplexityNormal,
		apistructs.IssueComplexityEasy,
	} {
		stacks = append(stacks, Stack{
			Name:  "",
			Value: i.GetZhName(),
		}) // TODO
	}
	return stacks
}

func (h *SourceStackHandler) GetStackColors() []string {
	return []string{"red", "yellow", "green"}
}

func (h *SourceStackHandler) GetIndexer() func(issue interface{}) string {
	return func(issue interface{}) string {
		switch issue.(type) {
		case *dao.IssueItem:
			return issue.(*dao.IssueItem).Complexity.GetZhName() // TODO
		case *model.LabelIssueItem:
			return issue.(*model.LabelIssueItem).Bug.Complexity.GetZhName()
		default:
			return ""
		}
	}
}
