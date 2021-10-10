package common

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func FixEmptyWord(em string) string {
	if em == "" {
		return "无"
	}
	return em
}

func GroupToPieData(issueList []dao.IssueItem, g func (issue *dao.IssueItem) string) []opts.PieData {
	counter := make(map[string]int)

	for _, i := range issueList {
		if i.Type != apistructs.IssueTypeBug {
			continue
		}
		counter[FixEmptyWord(g(&i))] ++
	}

	var data []opts.PieData
	for k, v := range counter {
		if v == 0 {
			continue
		}
		data = append(data, opts.PieData{
			Name: k,
			Value: v,
			Label: &opts.Label{
				Formatter: PieChartFormat,
				Show: true,
			},
		})
	}

	return data
}