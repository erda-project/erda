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

package burnoutChart

import (
	"context"
	"encoding/json"
	"strconv"
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
	base.InitProviderWithCreator(common.ScenarioKeyTestDashboard, "burnoutChart", func() servicehub.Provider {
		return &BurnoutChart{}
	})
}

func (f *BurnoutChart) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := f.InitFromProtocol(ctx, c); err != nil {
		return err
	}

	h := gshelper.NewGSHelper(gs)
	f.Issues = h.GetIssueList()

	dates := make([]time.Time, 0)
	dateMap := make(map[time.Time]int)
	itr := h.GetIteration()
	if itr.StartedAt == nil || itr.FinishedAt == nil {
		return nil
	}
	for rd := common.RangeDate(*itr.StartedAt, *itr.FinishedAt); ; {
		date := rd()
		if date.IsZero() {
			break
		}
		dates = append(dates, date)
		dateMap[date] = 0
	}

	issueCreateMap := make(map[time.Time][]dao.IssueItem, 0)
	issueFinishMap := make(map[time.Time][]dao.IssueItem, 0)
	for _, issue := range f.Issues {
		if h.GetBurnoutChartType() == "requirement" &&
			issue.Type != apistructs.IssueTypeRequirement {
			continue
		}
		if h.GetBurnoutChartType() == "task" &&
			issue.Type != apistructs.IssueTypeTask {
			continue
		}
		issueCreateMap[common.DateTime(issue.CreatedAt)] = append(issueCreateMap[common.DateTime(issue.CreatedAt)], issue)
		if issue.FinishTime != nil {
			issueFinishMap[common.DateTime(*issue.FinishTime)] = append(issueFinishMap[common.DateTime(*issue.FinishTime)], issue)
		}
	}

	cur := 0
	sum := 0
	for i, date := range dates {
		if date.After(time.Now()) {
			cur = i
			break
		}
		if _, ok := issueCreateMap[date]; ok {
			if h.GetBurnoutChartDimension() == "total" {
				sum += len(issueCreateMap[date])
			}
			if h.GetBurnoutChartDimension() == "workTime" {
				for _, issue := range issueCreateMap[date] {
					if issue.ManHour != "" {
						var manHour apistructs.IssueManHour
						if err := json.Unmarshal([]byte(issue.ManHour), &manHour); err != nil {
							return err
						}
						sum += int(manHour.EstimateTime) / 60
					}
				}
			}
		}

		if _, ok := issueFinishMap[date]; ok {
			if h.GetBurnoutChartDimension() == "total" {
				sum -= len(issueFinishMap[date])
			}
			if h.GetBurnoutChartDimension() == "workTime" {
				for _, issue := range issueFinishMap[date] {
					if issue.ManHour != "" {
						var manHour apistructs.IssueManHour
						if err := json.Unmarshal([]byte(issue.ManHour), &manHour); err != nil {
							return err
						}
						sum -= int(manHour.ThisElapsedTime) / 60
					}
				}
			}
		}
		dateMap[date] = sum
	}

	f.Type = "Chart"
	f.Props = Props{
		ChartType: "line",
		Title:     "燃尽图",
		PureChart: true,
		Option: Option{
			XAxis: XAxis{
				Type: "category",
				Data: func() []string {
					ss := make([]string, 0, len(dates))
					for _, v := range dates {
						ss = append(ss, v.Format("01-02"))
					}
					return ss
				}(),
			},
			YAxis: YAxis{
				Type: "value",
				AxisLine: map[string]interface{}{
					"lineStyle": map[string]interface{}{
						"color": "rgba(48,38,71,0.30)",
					},
				},
				AxisLabel: map[string]interface{}{
					"formatter": func() string {
						if h.GetBurnoutChartDimension() == "total" {
							return "{value} 个"
						}
						return "{value} h"
					}(),
				},
			},
			Legend: Legend{
				Show:   true,
				Bottom: true,
				Data:   []string{"实际燃尽"},
			},
			Tooltip: map[string]interface{}{
				"trigger": "axis",
			},
			Series: []Series{
				{
					Data: func() []int {
						counts := make([]int, 0)
						for _, v := range dates[:cur] {
							counts = append(counts, dateMap[v])
						}
						return counts
					}(),
					Name:   "实际燃尽",
					Type:   "line",
					Smooth: false,
					ItemStyle: map[string]interface{}{
						"color": "#D84B65",
					},
					MarkLine: MarkLine{
						Label: map[string]interface{}{"position": "middle"},
						LineStyle: map[string]interface{}{
							"color": "rgba(48,38,71,0.20)",
						},
						Data: [][]Data{
							{
								{
									Name: "预设燃尽",
									Coord: func() []string {
										ss := make([]string, 0, 2)
										ss = append(ss, dates[0].Format("01-02"), strconv.Itoa(dateMap[dates[0]]))
										return ss
									}(),
								},
								{
									Coord: func() []string {
										ss := make([]string, 0, 2)
										ss = append(ss, dates[len(dates)-1].Format("01-02"), "0")
										return ss
									}(),
								},
							},
						},
					},
				},
			},
		},
	}

	return f.SetToProtocolComponent(c)
}
