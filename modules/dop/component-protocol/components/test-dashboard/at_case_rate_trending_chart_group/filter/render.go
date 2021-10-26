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

package at_case_rate_trending_chart_filter

import (
	"context"
	"encoding/json"
	"time"

	"github.com/recallsong/go-utils/container/slice"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common/gshelper"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

func init() {
	base.InitProviderWithCreator(common.ScenarioKeyTestDashboard, "at_case_rate_trending_chart_filter", func() servicehub.Provider {
		return &Filter{}
	})
}

func (f *Filter) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	h := gshelper.NewGSHelper(gs)

	if err := f.initFromProtocol(ctx, c); err != nil {
		return err
	}
	times := f.State.Values.Time
	if len(times) != 2 {
		times = []int64{time.Now().AddDate(0, 0, -7).Unix() * 1000, time.Now().Unix() * 1000}
		f.State.Values.Time = times
	}
	timeStart := time.Unix(times[0]/1000, 0).Format("2006-01-02 15:04:05")
	timeEnd := time.Unix(times[1]/1000, 0).Format("2006-01-02 15:04:05")

	h.SetAtCaseRateTrendingTimeFilter(gshelper.AtSceneAndApiTimeFilter{
		TimeStart: timeStart,
		TimeEnd:   timeEnd,
	})

	if err := f.setState(ctx, gs); err != nil {
		return err
	}

	return f.setToComponent(c)
}

func (f *Filter) setState(ctx context.Context, gs *cptype.GlobalStateData) error {
	h := gshelper.NewGSHelper(gs)
	globalAtPlans := h.GetGlobalAutoTestPlanList()
	selectedItrsByID := h.GetGlobalSelectedIterationsByID()
	f.State.Conditions = []filter.PropCondition{
		{
			EmptyText: cputil.I18n(ctx, "all"),
			Fixed:     true,
			Key:       "atPlanIDs",
			Label:     cputil.I18n(ctx, "Test Plan"),
			ShowIndex: 1,
			Options: func() (opts []filter.PropConditionOption) {
				type Obj struct {
					filter.PropConditionOption
					itrCreateTime time.Time
				}
				var beforeOpts []Obj // used for order
				for _, plan := range globalAtPlans {
					itrPrefix := cputil.I18n(ctx, "[${iteration}: %s]", plan.IterationName)
					if plan.IterationID <= 0 {
						itrPrefix = cputil.I18n(ctx, "[${no-iteration}]")
					}
					beforeOpts = append(beforeOpts, Obj{
						PropConditionOption: filter.PropConditionOption{
							Label: cputil.I18n(ctx, "%s %s", itrPrefix, plan.Name),
							Value: plan.ID,
						},
						itrCreateTime: func() time.Time {
							itr, ok := selectedItrsByID[plan.IterationID]
							if !ok {
								return time.Time{}
							}
							return itr.CreatedAt
						}(),
					})
				}
				slice.Sort(beforeOpts, func(i, j int) bool {
					return beforeOpts[i].itrCreateTime.After(beforeOpts[j].itrCreateTime)
				})
				for _, bo := range beforeOpts {
					opts = append(opts, bo.PropConditionOption)
				}
				return
			}(),
			Type: filter.PropConditionTypeSelect,
		},
		{
			CustomProps: func() map[string]interface{} {
				now := time.Now()
				weekAgo := now.AddDate(0, 0, -7)
				monthAgo := now.AddDate(0, -1, 0)

				customProps := CustomProps{
					AllowClear: false,
					Ranges: Ranges{
						Week:  []int64{weekAgo.Unix() * 1000, now.Unix() * 1000},
						Month: []int64{monthAgo.Unix() * 1000, now.Unix() * 1000},
					},
				}

				b, _ := json.Marshal(&customProps)
				customPropsMap := make(map[string]interface{}, 0)
				_ = json.Unmarshal(b, &customPropsMap)
				return customPropsMap
			}(),
			Label:     cputil.I18n(ctx, "time"),
			Type:      filter.PropConditionTypeRangePicker,
			Fixed:     true,
			ShowIndex: 2,
			Key:       "time",
		},
	}
	h.SetRateTrendingFilterTestPlanList(func() (selectedAtPlans []apistructs.TestPlanV2) {
		// not selected, return all
		if len(f.State.Values.AtPlanIDs) == 0 {
			return globalAtPlans
		}
		globalAtPlanMap := make(map[uint64]apistructs.TestPlanV2)
		for _, plan := range globalAtPlans {
			globalAtPlanMap[plan.ID] = plan
		}
		for _, planID := range f.State.Values.AtPlanIDs {
			selectedAtPlans = append(selectedAtPlans, globalAtPlanMap[planID])
		}
		return
	}())

	return nil
}
