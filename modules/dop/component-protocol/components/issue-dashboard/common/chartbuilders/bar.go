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

package chartbuilders

import (
	"sort"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/erda-project/erda-infra/pkg/strutil"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/stackhandlers"
)

type BarBuilder struct {
	Items        []interface{}
	StackHandler stackhandlers.StackHandler
	FixedXAxisOrTop
	YAxisOpt
	StackOpt
	DataHandleOpt
	Result
}

type FixedXAxisOrTop struct {
	XAxis    []string
	Top      int
	XIndexer func(interface{}) string

	MaxValue          int
	XDisplayConverter func(opt *FixedXAxisOrTop) opts.XAxis
}

type YAxisOpt struct {
	YAxis             []string
	YDisplayConverter func(opt *YAxisOpt) opts.YAxis
}

type StackOpt struct {
	EnableSum bool
	SkipEmpty bool
}

type DataHandleOpt struct {
	SeriesConverter func(name string, data []*int) charts.SingleSeries
	DataWhiteList   []string
	FillZero        bool
}

type Result struct {
	Bar  *charts.Bar
	Size int
}

func (f *BarBuilder) Generate() error {
	series, colors, realY, sum := f.groupToBarData()

	bar := charts.NewBar()
	bar.Colors = colors
	bar.MultiSeries = series

	if f.XDisplayConverter != nil {
		maxValue := 0
		for _, t := range sum {
			if t > maxValue {
				maxValue = t
			}
		}
		//maxValue = int(float32(maxValue) * 1.2)
		f.FixedXAxisOrTop.MaxValue = maxValue
		bar.XAxisList[0] = f.XDisplayConverter(&f.FixedXAxisOrTop)
	}

	if f.YDisplayConverter != nil {
		f.YAxisOpt.YAxis = realY
		bar.YAxisList[0] = f.YDisplayConverter(&f.YAxisOpt)
	}

	//if f.Top > 0 {
	//	pageSize := 10
	//	lastIdx := float32(f.Size-1)
	//	if f.Size > pageSize {
	//		bar.DataZoomList = []opts.DataZoom{
	//			{Type: "slider", Start: lastIdx - float32(pageSize), End: lastIdx, YAxisIndex: 0},
	//		}
	//	}
	//}

	f.Bar = bar
	f.Size = len(sum)
	return nil
}

func (f *BarBuilder) groupToBarData() (charts.MultiSeries, []string, []string, []int) {
	counter := make(map[string]map[string]int)
	counterSingle := make(map[string]int)

	stackIndexer := f.StackHandler.GetIndexer()

	for _, i := range f.Items {
		y := FixEmptyWord(stackIndexer(i))
		if !showStack(f.DataWhiteList, y) {
			continue
		}
		x := FixEmptyWord(f.XIndexer(i))
		if _, ok := counter[y]; !ok {
			counter[y] = make(map[string]int)
		}
		counter[y][x]++
		counterSingle[x]++
	}

	var ms charts.MultiSeries
	xl := len(f.XAxis)
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
		last := f.Top
		xAxis := make([]string, 0)
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
		f.XAxis = xAxis
	}
	var colors []string
	total := make([]int, len(f.XAxis))

	for _, stack := range f.StackHandler.GetStacks() {
		if !showStack(f.DataWhiteList, stack.Value) {
			continue
		}
		rowData := make([]*int, xl)
		for i, x := range f.XAxis {
			v := counter[stack.Value][x]
			if v > 0 || !f.SkipEmpty {
				rowData[i] = &v
			}
			total[i] += v
		}
		ms = append(ms, f.SeriesConverter(stack.Name, rowData))
		colors = append(colors, stack.Color)
	}

	if f.EnableSum && len(colors) > 1 {
		totalRes := make([]*int, len(total))
		for i := range total {
			if !f.SkipEmpty {
				totalRes[i] = &total[i]
			}
		}

		msNew := make(charts.MultiSeries, len(ms)+1)
		colorsNew := make([]string, len(colors)+1)

		msNew[0] = f.SeriesConverter("全部", totalRes)
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

	return ms, colors, f.XAxis, total
}

func showStack(wl []string, item string) bool {
	if len(wl) == 0 { // empty means not filter
		return true
	}
	return strutil.Exist(wl, item)
}

func GetStackBarSingleSeriesConverter() func(name string, data []*int) charts.SingleSeries {
	return func(name string, data []*int) charts.SingleSeries {
		return charts.SingleSeries{
			Name:  name,
			Stack: "total",
			Data:  data,
			Label: &opts.Label{
				//Formatter: "{a}:{c}",
				Show: true,
			},
			BarGap: "30%",
		}
	}
}
