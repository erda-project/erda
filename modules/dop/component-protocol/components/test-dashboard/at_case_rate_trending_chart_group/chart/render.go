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

package chart

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common/gshelper"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator(common.ScenarioKeyTestDashboard, "at_case_rate_trending_chart", func() servicehub.Provider {
		return &Chart{}
	})
}

type Chart struct {
	base.DefaultProvider

	PData []Data `json:"pData"`
	EData []Data `json:"eData"`
	XAxis XAxis  `json:"xAxis"`
}

type Props struct {
	ChartType string `json:"chartType"`
	Option    Option `json:"option"`
	Title     string `json:"title"`
}

type Data struct {
	Value string `json:"value"`
}

type Option struct {
	Legend struct {
		Show bool `json:"show"`
	} `json:"legend"`
	Series []Series `json:"series"`
	XAxis  XAxis    `json:"xAxis"`
	YAxis  YAxis    `json:"yAxis"`
}

type Series struct {
	AreaStyle struct {
		Opacity float64 `json:"opacity"`
	} `json:"areaStyle"`
	Data  []Data `json:"data"`
	Label struct {
		Show bool `json:"show"`
	} `json:"label"`
	Name string `json:"name"`
}
type XAxis struct {
	Data []string `json:"data"`
}

type AxisLabel struct {
	Formatter string `json:"formatter"`
}

type YAxis struct {
	AxisLabel AxisLabel `json:"axisLabel"`
}

func (ch *Chart) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	h := gshelper.NewGSHelper(gs)

	historyList := h.GetAtTestPlanExecHistoryList()

	pData := make([]Data, 0, len(historyList))
	eData := make([]Data, 0, len(historyList))
	xAxis := make([]string, 0, len(historyList))
	var sucApiNum, execApiNum, totalApiNum int64
	for _, v := range historyList {
		if v.Type != apistructs.AutoTestPlan {
			continue
		}
		sucApiNum += v.SuccessApiNum
		execApiNum += v.ExecuteApiNum
		totalApiNum += v.TotalApiNum
		pData = append(pData, Data{
			Value: calRate(sucApiNum, totalApiNum),
		})
		eData = append(eData, Data{
			Value: calRate(execApiNum, totalApiNum),
		})
		xAxis = append(xAxis, v.ExecuteTime.Format("2006-01-02 15:04:05"))
	}
	ch.EData = eData
	ch.PData = pData
	ch.XAxis = XAxis{xAxis}
	c.Props = ch.convertToProps()
	return nil
}

func calRate(num, numTotal int64) string {
	if numTotal == 0 {
		return "0"
	}
	return fmt.Sprintf("%.2f", float64(num)/float64(numTotal)*100)
}

func (ch *Chart) convertToProps() Props {
	return Props{
		ChartType: "line",
		Title:     "",
		Option: Option{
			Legend: struct {
				Show bool `json:"show"`
			}{Show: true},
			Series: []Series{
				{
					AreaStyle: struct {
						Opacity float64 `json:"opacity"`
					}{Opacity: 0.1},
					Data: ch.PData,
					Label: struct {
						Show bool `json:"show"`
					}{Show: true},
					Name: "通过率",
				},
				{
					AreaStyle: struct {
						Opacity float64 `json:"opacity"`
					}{Opacity: 0.1},
					Data: ch.EData,
					Label: struct {
						Show bool `json:"show"`
					}{Show: true},
					Name: "执行率",
				},
			},
			XAxis: ch.XAxis,
			YAxis: YAxis{AxisLabel: AxisLabel{
				Formatter: "{value}%",
			}},
		},
	}
}
