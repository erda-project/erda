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

package stateVerticalBarChart

import (
	"context"
	"encoding/json"
	"github.com/go-echarts/go-echarts/v2/charts"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("issue-dashboard", "stateVerticalBarChart",
		func() servicehub.Provider { return &ComponentAction{} })
}

func (f *ComponentAction) getState(c *cptype.Component) error {
	d, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	return json.Unmarshal(d, &f.State)
}

func (f *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {

	if err := f.getState(c); err != nil {
		return err
	}

	stateMap := make(map[uint64]dao.IssueState)
	for _, i := range f.State.IssueStateList {
		stateMap[i.ID] = i
	}

	bar := charts.NewBar()
	bar.Colors = []string{"green", "blue", "orange", "red"}
	var yAxis []string
	for _, i := range apistructs.IssuePriorityList { // TODO: should changeable
		yAxis = append(yAxis, i.GetZhName())
	}
	bar.Legend.Data = yAxis
	// x is always stable
	var xAxis []string
	for _, i := range f.State.IssueStateList {
		xAxis = append(xAxis, i.Name)
		//bar.ExtendXAxis(opts.XAxis{
		//	Name:        i.Name,
		//})
	}
	bar.SetXAxis(xAxis)

	bar.MultiSeries = common.GroupToVerticalBarData(f.State.IssueList, yAxis, xAxis, func(issue *dao.IssueItem) string {
		return stateMap[uint64(issue.State)].Name
	}, func(issue *dao.IssueItem) string {
		return issue.Priority.GetZhName()
	})
	bar.Tooltip.Show = true

	props := make(map[string]interface{})
	props["title"] = "缺陷状态分布" // TODO: change by filter
	props["chartType"] = "bar"
	props["option"] = bar.JSON()

	c.Props = props
	return nil
}
