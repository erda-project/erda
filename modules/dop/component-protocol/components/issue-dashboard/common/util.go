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

package common

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
)

func FixEmptyWord(em string) string {
	if em == "" {
		return "无"
	}
	return em
}

func GroupToPieData(issueList []dao.IssueItem, g func(issue *dao.IssueItem) string) []opts.PieData {
	counter := make(map[string]int)

	for _, i := range issueList {
		if i.Type != apistructs.IssueTypeBug {
			continue
		}
		counter[FixEmptyWord(g(&i))]++
	}

	var data []opts.PieData
	for k, v := range counter {
		if v == 0 {
			continue
		}
		data = append(data, opts.PieData{
			Name:  k,
			Value: v,
			Label: &opts.Label{
				Formatter: PieChartFormat,
				Show:      true,
			},
		})
	}

	return data
}

func GroupToVerticalBarData(issueList []dao.IssueItem, yAxis []string, xAxis []string, yIdx func(issue *dao.IssueItem) string, xIdx func(issue *dao.IssueItem) string) charts.MultiSeries {
	counter := make(map[string]map[string]int)

	for _, i := range issueList {
		if i.Type != apistructs.IssueTypeBug {
			continue
		}
		y := FixEmptyWord(yIdx(&i))
		x := FixEmptyWord(xIdx(&i))
		if _, ok := counter[y]; !ok {
			counter[y] = make(map[string]int)
		}
		counter[y][x]++
	}

	var ms charts.MultiSeries
	for _, y := range yAxis {
		var rowData []int
		for _, x := range xAxis {
			rowData = append(rowData, counter[y][x])
		}
		ms = append(ms, charts.SingleSeries{
			Name:  y,
			Data:  rowData,
			Stack: "total",
			Label: &opts.Label{
				Formatter: "{a}:{c}",
				Show:      true,
			},
		})
	}

	return ms
}

func IssueListRetriever(issues []dao.IssueItem, match func(i int) bool) []dao.IssueItem {
	res := make([]dao.IssueItem, 0)
	for i, issue := range issues {
		if match(i) {
			res = append(res, issue)
		}
	}
	return res
}
