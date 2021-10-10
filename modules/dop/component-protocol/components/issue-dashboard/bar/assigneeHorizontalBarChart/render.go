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

package assigneeHorizontalBarChart

import (
	"context"
	"encoding/json"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("issue-dashboard", "assigneeHorizontalBarChart",
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

	memberMap := make(map[string]*apistructs.Member)
	for i := range f.State.Members {
		m := &f.State.Members[i]
		memberMap[m.UserID] = m
	}

	var bugList []interface{}
	for i := range f.State.IssueList {
		issue := f.State.IssueList[i]
		if issue.Type != apistructs.IssueTypeBug {
			continue
		}
		bugList = append(bugList, &issue)
	}

	bar := charts.NewBar()
	bar.Colors = []string{"green", "blue", "orange", "red"}
	var yAxis []string
	for _, i := range apistructs.IssuePriorityList {
		yAxis = append(yAxis, i.GetZhName())
	}
	bar.XAxisList[0] = opts.XAxis{
		Type: "value",
	}

	var realY []string
	bar.MultiSeries, realY = common.GroupToVerticalBarData(bugList, yAxis, nil, func(issue interface{}) string {
		return issue.(*dao.IssueItem).Priority.GetZhName()
	}, func(issue interface{}) string {
		return issue.(*dao.IssueItem).Assignee
	}, func(name string, data []*int) charts.SingleSeries {
		return charts.SingleSeries{
			Name:  name,
			Stack: "total",
			Data:  data,
			Label: &opts.Label{
				Formatter: "{a}:{c}",
				Show:      true,
			},
		}
	}, 500)

	var nameY []string
	for _, userID := range realY {
		var name string
		m, ok := memberMap[userID]
		if ok && m != nil {
			name = m.Nick
		}
		nameY = append(nameY, name)
	}

	bar.YAxisList[0] = opts.YAxis{
		Type: "category",
		Data: nameY,
	}

	props := make(map[string]interface{})
	props["title"] = "未完成缺陷按处理人分布（TOP 500）"
	props["chartType"] = "bar"
	props["option"] = bar.JSON()

	c.Props = props
	return nil
}
