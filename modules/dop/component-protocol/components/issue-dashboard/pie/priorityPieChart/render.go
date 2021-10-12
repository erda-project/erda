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

package priorityPieChart

import (
	"context"
	"encoding/json"

	"github.com/go-echarts/go-echarts/v2/charts"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/stackhandlers"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("issue-dashboard", "priorityPieChart",
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

	helper := gshelper.NewGSHelper(gs)

	handler := stackhandlers.NewPriorityStackHandler(false)

	seriesData, colors := common.GroupToPieData(helper.GetIssueList(), handler)

	pie := charts.NewPie()
	pie.Tooltip.Show = true
	pie.Tooltip.Trigger = "item"
	pie.Colors = colors
	pie.AddSeries("优先级", seriesData, common.GetPieSeriesOpt())
	props := make(map[string]interface{})
	props["title"] = "按优先级"
	props["chartType"] = "pie"
	props["option"] = pie.JSON()

	c.Props = props
	c.State = nil
	return nil
}
