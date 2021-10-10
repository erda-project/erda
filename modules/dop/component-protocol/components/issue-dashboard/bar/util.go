package bar

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type PriorityStackHandler struct {
}

func (h PriorityStackHandler) GetStacks() []string {
	var stacks []string
	for _, i := range apistructs.IssuePriorityList {
		stacks = append(stacks, i.GetZhName())
	}
	return stacks
}

func (h PriorityStackHandler) GetStackColors() []string {
	return []string{"green", "blue", "orange", "red"}
}

func (h PriorityStackHandler) GetIndexer() func(issue interface{}) string {
	return func(issue interface{}) string {
		return issue.(*dao.IssueItem).Priority.GetZhName()
	}
}


type LabelPriorityStackHandler struct {
	BugMap map[uint64]*dao.IssueItem
}

func (h LabelPriorityStackHandler) GetStacks() []string {
	var stacks []string
	for _, i := range apistructs.IssuePriorityList {
		stacks = append(stacks, i.GetZhName())
	}
	return stacks
}

func (h LabelPriorityStackHandler) GetStackColors() []string {
	return []string{"green", "blue", "orange", "red"}
}

func (h LabelPriorityStackHandler) GetIndexer() func(issue interface{}) string {
	return func(label interface{}) string {
		l := label.(*dao.IssueLabel)
		if l == nil {
			return ""
		}
		bug, ok := h.BugMap[l.RefID]
		if !ok {
			return ""
		}
		return bug.Priority.GetZhName()
	}
}

func GetAssigneeIndexer() func(issue interface{}) string {
	return func(issue interface{}) string {
		return issue.(*dao.IssueItem).Assignee
	}
}

func GetHorizontalStackBarSingleSeriesConverter() func(name string, data []*int) charts.SingleSeries {
	return func(name string, data []*int) charts.SingleSeries {
		return charts.SingleSeries{
			Name:  name,
			Stack: "total",
			Data:  data,
			Label: &opts.Label{
				Formatter: "{a}:{c}",
				Show:      true,
			},
		}
	}
}
