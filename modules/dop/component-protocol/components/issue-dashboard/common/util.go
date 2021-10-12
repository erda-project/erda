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
	"sort"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/stackhandlers"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/pkg/strutil"
)

func FixEmptyWord(em string) string {
	if em == "" {
		return "无"
	}
	return em
}

func GroupToPieData(issueList []dao.IssueItem, stackHandler stackhandlers.StackHandler) ([]opts.PieData, []string) {
	counter := make(map[string]int)
	indexer := stackHandler.GetIndexer()

	for _, i := range issueList {
		if i.Type != apistructs.IssueTypeBug {
			continue
		}
		counter[FixEmptyWord(indexer(&i))]++
	}

	var data []opts.PieData
	var colors []string
	for _, stack := range stackHandler.GetStacks() {
		cnt := counter[stack.Value]
		if cnt <= 0 {
			continue
		}
		data = append(data, opts.PieData{
			Name:  stack.Name,
			Value: cnt,
			Label: &opts.Label{
				Formatter: PieChartFormat,
				Show:      true,
			},
		})
		colors = append(colors, stack.Color)
	}
	return data, colors
}

func GroupToBarData(itemList []interface{}, wl []string, stackHandler stackhandlers.StackHandler, xAxis []string,
	xIdx func(issue interface{}) string,
	seriesConverter func(name string, data []*int) charts.SingleSeries, top int, enableTotal bool, skipEmpty bool) (charts.MultiSeries, []string, []string, []int) {
	counter := make(map[string]map[string]int)
	counterSingle := make(map[string]int)

	stackIndexer := stackHandler.GetIndexer()

	for _, i := range itemList {
		y := FixEmptyWord(stackIndexer(i))
		if !showStack(wl, y) {
			continue
		}
		x := FixEmptyWord(xIdx(i))
		if _, ok := counter[y]; !ok {
			counter[y] = make(map[string]int)
		}
		counter[y][x]++
		counterSingle[x]++
	}

	var ms charts.MultiSeries
	xl := len(xAxis)
	if xl == 0 {
		// use top instead
		var sl counterList
		for k, v := range counterSingle {
			if v == 0 {
				continue
			}
			sl = append(sl, counterItem{Name: k, Value: v})
		}
		sort.Sort(sl)
		last := top
		xAxis = make([]string, 0)
		xl = 0 // reset
		for _, item := range sl {
			if last <= 0 || item.Value <= 0 {
				break
			}
			xAxis = append(xAxis, item.Name)
			xl++
			last--
		}
		// reverse for top
		for i, j := 0, len(xAxis)-1; i < j; i, j = i+1, j-1 {
			xAxis[i], xAxis[j] = xAxis[j], xAxis[i]
		}
	}
	var colors []string
	total := make([]int, len(xAxis))

	for _, stack := range stackHandler.GetStacks() {
		if !showStack(wl, stack.Value) {
			continue
		}
		rowData := make([]*int, xl)
		for i, x := range xAxis {
			v := counter[stack.Value][x]
			if v > 0 || !skipEmpty {
				rowData[i] = &v
			}
			total[i] += v
		}
		ms = append(ms, seriesConverter(stack.Name, rowData))
		colors = append(colors, stack.Color)
	}

	if enableTotal && len(colors) > 1 {
		totalRes := make([]*int, len(total))
		for i := range total {
			if !skipEmpty {
				totalRes[i] = &total[i]
			}
		}

		msNew := make(charts.MultiSeries, len(ms)+1)
		colorsNew := make([]string, len(colors)+1)

		msNew[0] = seriesConverter("全部", totalRes)
		colorsNew[0] = "gray"
		for i := range ms {
			msNew[i+1] = ms[i]
		}
		ms = msNew
		for i := range colors {
			colorsNew[i+1] = colors[i]
		}
		colors = colorsNew
	}

	return ms, colors, xAxis, total
}

func showStack(wl []string, item string) bool {
	if len(wl) == 0 { // empty means not filter
		return true
	}
	return strutil.Exist(wl, item)
}

func GetAssigneeIndexer() func(issue interface{}) string {
	return func(issue interface{}) string {
		return issue.(*dao.IssueItem).Assignee
	}
}

func GetStackBarSingleSeriesConverter() func(name string, data []*int) charts.SingleSeries {
	return func(name string, data []*int) charts.SingleSeries {
		return charts.SingleSeries{
			Name:  name,
			Stack: "total",
			Data:  data,
			Label: &opts.Label{
				//Formatter: "{a}:{c}",
				Show:      true,
			},
		}
	}
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

type counterItem struct {
	Name  string
	Value int
}

type counterList []counterItem

func (l counterList) Less(i, j int) bool { return l[i].Value > l[j].Value }
func (l counterList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l counterList) Len() int           { return len(l) }
