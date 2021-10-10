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
)

func FixEmptyWord(em string) string {
	if em == "" {
		return "无"
	}
	return em
}

func GroupToPieData(issueList []dao.IssueItem, stackHandler stackhandlers.StackHandler) []opts.PieData {
	counter := make(map[string]int)
	indexer := stackHandler.GetIndexer()

	for _, i := range issueList {
		if i.Type != apistructs.IssueTypeBug {
			continue
		}
		counter[FixEmptyWord(indexer(&i))]++
	}

	var data []opts.PieData
	for _, stack := range stackHandler.GetStacks() {
		cnt := counter[stack]
		if cnt <= 0 {
			continue
		}
		data = append(data, opts.PieData{
			Name:  stack,
			Value: cnt,
			Label: &opts.Label{
				Formatter: PieChartFormat,
				Show:      true,
			},
		})
	}
	return data
}

func GroupToVerticalBarData(itemList []interface{}, stackHandler stackhandlers.StackHandler, xAxis []string,
	xIdx func(issue interface{}) string,
	seriesConverter func(name string, data []*int) charts.SingleSeries, top int) (charts.MultiSeries, []string) {
	counter := make(map[string]map[string]int)
	counterSingle := make(map[string]int)

	stackIndexer := stackHandler.GetIndexer()

	for _, i := range itemList {
		y := FixEmptyWord(stackIndexer(i))
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
	}
	for _, stack := range stackHandler.GetStacks() {
		rowData := make([]*int, xl)
		for i, x := range xAxis {
			v := counter[stack][x]
			if v > 0 {
				rowData[i] = &v
			}
		}
		ms = append(ms, seriesConverter(stack, rowData))
	}

	return ms, xAxis
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

func IssueListRetriever(issues []dao.IssueItem, match func(i int) bool) []dao.IssueItem {
	res := make([]dao.IssueItem, 0)
	for i, issue := range issues {
		if match(i) {
			res = append(res, issue)
		}
	}
	return res
}

func IssueListFilter(issues []dao.IssueItem, match func(i int) bool) []interface{} {
	res := make([]interface{}, 0)
	for i := range issues {
		if match(i) {
			res = append(res, &issues[i])
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

func StackRetriever(t string) stackhandlers.StackHandler {
	var handler stackhandlers.StackHandler
	switch t {
	case Priority:
		handler = stackhandlers.PriorityStackHandler{}
	case Complexity:
		handler = stackhandlers.ComplexityStackHandler{}
	case Severity:
		handler = stackhandlers.SeverityStackHandler{}
	}
	return handler
}
