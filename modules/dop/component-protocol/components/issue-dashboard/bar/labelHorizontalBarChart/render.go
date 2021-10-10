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

package labelHorizontalBarChart

import (
	"context"
	"encoding/json"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/go-echarts/go-echarts/v2/charts"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("issue-dashboard", "labelHorizontalBarChart",
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
	var leg []string
	for _, i := range f.State.IssueStateList {
		leg = append(leg, i.Name)
	}
	bar.Legend.Data = leg
	bar.AddSeries("缺陷分布", []opts.BarData{

	}, func(s *charts.SingleSeries) {
		s.Animation = true
	})
	props := make(map[string]interface{})
	props["title"] = "缺陷状态分布" // TODO: change by filter
	props["chartType"] = "bar"
	props["option"] = bar.JSON()

	c.Props = props
	return nil
}
