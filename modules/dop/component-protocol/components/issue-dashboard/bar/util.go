package bar

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
)

type LabelIssueItem struct {
	LabelRel *dao.IssueLabel
	Bug      *dao.IssueItem
}

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
		switch issue.(type) {
		case *dao.IssueItem:
			return issue.(*dao.IssueItem).Priority.GetZhName()
		case *LabelIssueItem:
			return issue.(*LabelIssueItem).Bug.Priority.GetZhName()
		default:
			return ""
		}
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
