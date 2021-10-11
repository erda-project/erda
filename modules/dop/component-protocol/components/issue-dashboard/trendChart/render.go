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

package trendChart

import (
	"context"
	"encoding/json"
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/pkg/strutil"
)

func init() {
	base.InitProviderWithCreator("issue-dashboard", "trendChart",
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

	if f.State.Iterations != nil {
		f.ChartDataRetriever(f.State.Iterations[0])
	}
	f.State.IssueList = nil
	return f.SetToProtocolComponent(c)
}

func rangeDate(start, end time.Time) func() time.Time {
	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, start.Location())
	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, end.Location())

	return func() time.Time {
		if start.After(end) {
			return time.Time{}
		}
		date := start
		start = start.AddDate(0, 0, 1)
		return date
	}
}

func (f *ComponentAction) ChartDataRetriever(iteration apistructs.Iteration) {
	dates := make([]string, 0)
	cMap := make(map[time.Time][]int)
	for rd := rangeDate(*iteration.StartedAt, *iteration.FinishedAt); ; {
		date := rd()
		if date.IsZero() {
			break
		}
		cMap[date] = []int{0, 0, 0}
	}

	first, last := [3]int{0, 0, 0}, [3]int{0, 0, 0}

	issues := common.IssueListRetriever(f.State.IssueList, func(i int) bool {
		v := f.State.IssueList[i].FilterPropertyRetriever(f.State.Values.Type)
		return f.State.Values.Value == nil || strutil.Exist(f.State.Values.Value, v)
	})
	for _, i := range issues {
		created := time.Date(i.CreatedAt.Year(), i.CreatedAt.Month(), i.CreatedAt.Day(), 0, 0, 0, 0, i.CreatedAt.Location())
		if created.Before(*iteration.StartedAt) {
			first[0] += 1
		} else if created.After(*iteration.FinishedAt) {
			last[0] += 1
		} else {
			cMap[created][0] += 1
		}

		if i.FinishTime != nil {
			closed := time.Date(i.FinishTime.Year(), i.FinishTime.Month(), i.FinishTime.Day(), 0, 0, 0, 0, i.FinishTime.Location())
			if closed.Before(*iteration.StartedAt) {
				first[1] += 1
			} else if created.After(*iteration.FinishedAt) {
				last[1] += 1
			} else {
				cMap[closed][1] += 1
			}
		}
	}

	newIssue, closedIssue, unClosedIssue := make([]int, 0), make([]int, 0), make([]int, 0)
	dates = append(dates, "更早")
	newIssue = append(newIssue, first[0])
	closedIssue = append(closedIssue, first[1])
	first[2] = first[0] - first[1]
	unClosedIssue = append(unClosedIssue, first[2])
	for rd := rangeDate(*iteration.StartedAt, *iteration.FinishedAt); ; {
		date := rd()
		if date.IsZero() {
			break
		}

		x := date.Format("1/2")
		dates = append(dates, x)
		newIssue = append(newIssue, cMap[date][0])
		closedIssue = append(closedIssue, cMap[date][1])
		unclose := unClosedIssue[len(unClosedIssue)-1] + cMap[date][0] - cMap[date][1]
		unClosedIssue = append(unClosedIssue, unclose)
	}

	dates = append(dates, "未来")
	newIssue = append(newIssue, last[0])
	closedIssue = append(closedIssue, last[1])
	last[2] = unClosedIssue[len(unClosedIssue)-1] + last[0] - last[1]
	unClosedIssue = append(unClosedIssue, last[2])

	f.Chart = common.Chart{
		Props: common.Props{
			Title:     "缺陷新增、关闭、未关闭数走势",
			ChartType: "line",
			Option: common.Option{
				XAxis: common.XAxis{
					Data: dates,
				},
				Color: []string{"blue", "green", "red"},
				Series: []common.Item{
					{
						Name: "新增",
						Data: newIssue,
						AreaStyle: common.AreaStyle{
							Opacity: 0.1,
						},
					},
					{
						Name: "关闭",
						Data: closedIssue,
						AreaStyle: common.AreaStyle{
							Opacity: 0.1,
						},
					},
					{
						Name: "未关闭",
						Data: unClosedIssue,
						AreaStyle: common.AreaStyle{
							Opacity: 0.1,
						},
					},
				},
			},
		},
	}
}
