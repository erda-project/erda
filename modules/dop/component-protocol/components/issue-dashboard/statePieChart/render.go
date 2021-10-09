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

package statePieChart

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("issue-dashboard", "statePieChart",
		func() servicehub.Provider { return &ComponentAction{} })
}

func (f *ComponentAction) InitFromProtocol(ctx context.Context, c *cptype.Component) error {
	// component 序列化
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, f); err != nil {
		return err
	}

	// sdk
	f.sdk = cputil.SDK(ctx)
	f.bdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	// if err := f.setInParams(ctx); err != nil {
	// 	return err
	// }

	return nil
}

func (f *ComponentAction) SetToProtocolComponent(c *cptype.Component) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}
	c.State = nil
	return nil
}

func (f *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := f.InitFromProtocol(ctx, c); err != nil {
		return err
	}

	f.ChartDataRetriever()
	f.State.IssueList = nil
	return f.SetToProtocolComponent(c)
}

func indexOf(element string, data []apistructs.IssueStateBelong) int {
	for k, v := range data {
		if element == string(v) {
			return k
		}
	}
	return -1
}

func (f *ComponentAction) ChartDataRetriever() {
	// states := []apistructs.IssueStateBelong{apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking, apistructs.IssueStateBelongWontfix,
	// 	apistructs.IssueStateBelongReopen, apistructs.IssueStateBelongResloved, apistructs.IssueStateBelongClosed}

	// values := []int{0, 0, 0, 0, 0, 0, 0}
	// for i, issue := range f.State.IssueList {
	// 	values[indexOf(issue.)]
	// }
	f.PieChart = common.PieChart{
		Props: common.PieChartProps{
			Title:     "按缺陷状态",
			ChartType: "pie",
			Option: common.PieChartOption{
				Color: []string{"green", "blue", "orange", "red", "lime", "olive", "yellow"},
				Series: []common.PieChartItem{
					{
						Name: "缺陷状态",
						Data: []common.PieChartPart{
							{
								Name:  "待处理",
								Value: 12,
								Label: common.Label{
									Fortmatter: common.PieChartFormat,
								},
							},
							{
								Name:  "进行中",
								Value: 12,
								Label: common.Label{
									Fortmatter: common.PieChartFormat,
								},
							},
							{
								Name:  "已解决",
								Value: 10,
								Label: common.Label{
									Fortmatter: common.PieChartFormat,
								},
							},
							{
								Name:  "重新打开",
								Value: 8,
								Label: common.Label{
									Fortmatter: common.PieChartFormat,
								},
							},
						},
					},
				},
			},
		},
	}
}
