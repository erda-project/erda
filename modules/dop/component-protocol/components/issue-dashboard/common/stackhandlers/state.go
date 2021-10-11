package stackhandlers

import (
	"fmt"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/model"
	"github.com/erda-project/erda/modules/dop/dao"
)

type StateStackHandler struct {
	IssueStateList []dao.IssueState
}

func NewStateStackHandler(issueStateList []dao.IssueState) *StateStackHandler {
	return &StateStackHandler{
		IssueStateList: issueStateList,
	}
}

var colorMap = map[apistructs.IssueStateBelong][]string{
	// 待处理
	apistructs.IssueStateBelongOpen: {"yellow"},
	// 进行中
	apistructs.IssueStateBelongWorking: {"blue", "steelblue", "darkslategray", "darkslateblue"},
	// 已解决
	apistructs.IssueStateBelongResloved: {"green"},
	// 已完成
	apistructs.IssueStateBelongDone: {"green"},
	// 重新打开
	apistructs.IssueStateBelongReopen: {"red"},
	// 无需修复
	apistructs.IssueStateBelongWontfix: {"orange", "grey"},
	// 已关闭
	apistructs.IssueStateBelongClosed: {"darkseagreen"},
}

func (h *StateStackHandler) GetStacks() []Stack {
	var stacks []Stack
	for _, i := range h.IssueStateList {
		stacks = append(stacks, Stack{
			Name:  i.Name,
			Value: fmt.Sprintf("%d", i.ID),
		})
	}
	return stacks
}

func (h *StateStackHandler) GetStackColors() []string {
	var colors []string
	belongCounter := make(map[apistructs.IssueStateBelong]int)
	for _, i := range h.IssueStateList {
		color := colorMap[i.Belong][belongCounter[i.Belong] % len(colorMap[i.Belong])]
		colors = append(colors, color)
		belongCounter[i.Belong]++
	}
	return colors
}

func (h *StateStackHandler) GetIndexer() func(issue interface{}) string {
	return func(issue interface{}) string {
		switch issue.(type) {
		case *dao.IssueItem:
			return fmt.Sprintf("%d", issue.(*dao.IssueItem).State)
		case *model.LabelIssueItem:
			return fmt.Sprintf("%d", issue.(*model.LabelIssueItem).Bug.State)
		default:
			return ""
		}
	}
}
