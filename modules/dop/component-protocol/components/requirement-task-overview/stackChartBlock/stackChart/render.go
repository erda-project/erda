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

package stackChart

import (
	"context"
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/requirement-task-overview/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/requirement-task-overview/common/gshelper"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator(common.ScenarioKeyTestDashboard, "stackChart", func() servicehub.Provider {
		return &StackChart{}
	})
}

func (f *StackChart) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := f.InitFromProtocol(ctx, c); err != nil {
		return err
	}

	h := gshelper.NewGSHelper(gs)
	itr := h.GetIteration()
	if itr.StartedAt == nil || itr.FinishedAt == nil {
		return nil
	}
	f.Itr = itr
	f.setIssues(h)
	if err := f.setStatesCircusMap(); err != nil {
		return err
	}
	if err := f.setIssueStates(h); err != nil {
		return err
	}
	f.setDateMap()
	f.setProps()
	return f.SetToProtocolComponent(c)
}

func (f *StackChart) setProps() {
	f.Type = "Chart"
	series := make([]Series, 0, len(f.States))
	for _, v := range f.States {
		data := make([]int, 0)
		for _, date := range f.Dates {
			data = append(data, f.DateMap[date][v.ID])
		}
		series = append(series, Series{
			Data: func() []int {
				array := make([]int, 0)
				for _, date := range f.Dates {
					array = append(array, f.DateMap[date][v.ID])
				}
				return array
			}(),
			Name:   v.Name,
			Stack:  "总量",
			Type:   "line",
			Smooth: false,
			Symbol: "none",
			AreaStyle: map[string]interface{}{
				"opacity": 1,
			},
			LineStyle: map[string]interface{}{
				"width": 0,
			},
		})
	}
	f.Props = Props{
		ChartType: "line",
		Title:     "累积流图",
		PureChart: true,
		Option: Option{
			Grid: map[string]interface{}{"bottom": 60},
			XAxis: XAxis{
				Type: "category",
				Data: func() []string {
					ss := make([]string, 0, len(f.Dates))
					for _, v := range f.Dates {
						ss = append(ss, v.Format("01-02"))
					}
					return ss
				}(),
			},
			YAxis: YAxis{
				Type: "value",
				AxisLine: map[string]interface{}{
					"lineStyle": map[string]interface{}{
						"color": "rgba(48,38,71,0.20)",
					},
				},
				AxisLabel: map[string]interface{}{
					"formatter": "{value} 个",
				},
			},
			Legend: Legend{
				Show:   true,
				Bottom: true,
				Data: func() []string {
					ss := make([]string, 0, len(f.States))
					for _, v := range f.States {
						ss = append(ss, v.Name)
					}
					return ss
				}(),
			},
			Tooltip: map[string]interface{}{
				"trigger": "axis",
			},
			Color:  []string{},
			Series: series,
		},
	}
}

func (f *StackChart) setDateMap() {
	dateMap := make(map[time.Time]map[uint64]int)
	for rd := common.RangeDate(*f.Itr.StartedAt, *f.Itr.FinishedAt); ; {
		date := rd()
		if date.IsZero() {
			break
		}
		f.Dates = append(f.Dates, date)
		count := make(map[uint64]int, 0)
		for _, v := range f.States {
			count[v.ID] = 0
		}
		dateMap[date] = count
	}

	baseList := make([]dao.IssueStateCirculation, 0)
	for k, v := range f.StatesCircusMap {
		if !common.DateTime(k).After(f.Dates[0]) {
			baseList = append(baseList, v...)
		}
	}

	issueIDMap := make(map[uint64]struct{})
	for _, v := range f.Issues {
		issueIDMap[v.ID] = struct{}{}
	}

	baseCount := make(map[uint64]int, 0)
	for _, v := range baseList {
		if _, ok := issueIDMap[v.IssueID]; !ok {
			continue
		}
		if _, ok := dateMap[f.Dates[0]][v.StateFrom]; ok && v.StateFrom != 0 {
			baseCount[v.StateFrom] --
		}
		if _, ok := dateMap[f.Dates[0]][v.StateTo]; ok {
			baseCount[v.StateTo] ++
		}

	}
	dateMap[f.Dates[0]] = deepCopy(baseCount)

	for i := 1; i < len(f.Dates); i++ {
		if _, ok := f.StatesCircusMap[f.Dates[i]]; ok {
			for _, v := range f.StatesCircusMap[f.Dates[i]] {
				if _, ok2 := issueIDMap[v.IssueID]; !ok2 {
					continue
				}
				if _, ok3 := dateMap[f.Dates[i]][v.StateFrom]; ok3 && v.StateFrom != 0 {
					baseCount[v.StateFrom] --
				}
				if _, ok3 := dateMap[f.Dates[i]][v.StateTo]; ok3 {
					baseCount[v.StateTo] ++
				}
			}
		}
		dateMap[f.Dates[i]] = deepCopy(baseCount)
	}
	f.DateMap = dateMap
}

func (f *StackChart) setIssueStates(h *gshelper.GSHelper) error {
	t := apistructs.IssueTypeRequirement
	if h.GetStackChartType() == "task" {
		t = apistructs.IssueTypeTask
	}
	states, err := f.issueSvc.GetIssuesStatesByProjectID(f.InParams.ProjectID, t)
	if err != nil {
		return err
	}
	f.States = states
	return nil
}

func (f *StackChart) setIssues(h *gshelper.GSHelper) {
	for _, issue := range h.GetIssueList() {
		if h.GetStackChartType() == "requirement" &&
			issue.Type != apistructs.IssueTypeRequirement {
			continue
		}
		if h.GetStackChartType() == "task" &&
			issue.Type != apistructs.IssueTypeTask {
			continue
		}
		f.Issues = append(f.Issues, issue)
	}
}

func (f *StackChart) setStatesCircusMap() error {
	statesCircus, err := f.issueSvc.ListStatesCircusByProjectID(f.InParams.ProjectID)
	if err != nil {
		return err
	}

	statesCircusMap := make(map[time.Time][]dao.IssueStateCirculation, 0)
	for _, v := range statesCircus {
		statesCircusMap[common.DateTime(v.CreatedAt)] = append(statesCircusMap[common.DateTime(v.CreatedAt)], v)
	}
	f.StatesCircusMap = statesCircusMap
	return nil
}

func deepCopy(count map[uint64]int) map[uint64]int {
	newCount := make(map[uint64]int, 0)
	for k, v := range count {
		newCount[k] = v
	}
	return newCount
}

var Colors = []string{
	"primary8", "primary7", "primary6", "primary5", "primary4", "primary3", "primary2", "primary1",
	"warning8", "warning7", "warning6", "warning5", "warning4", "warning3", "warning2", "warning1",
}

func getColors(len int) []string {
	if len < 16 {
		return Colors[:len]
	}
	return Colors
}
