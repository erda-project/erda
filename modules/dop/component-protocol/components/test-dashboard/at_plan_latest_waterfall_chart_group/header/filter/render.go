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

package filter

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common/gshelper"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

func init() {
	base.InitProviderWithCreator(common.ScenarioKeyTestDashboard, "at_plan_latest_waterfall_chart_filter", func() servicehub.Provider {
		return &Filter{}
	})
}

type Filter struct {
	base.DefaultProvider

	State State `json:"state,omitempty"`
}

type State struct {
	Conditions []filter.PropCondition `json:"conditions,omitempty"`
	Values     PipelineIDValues       `json:"values,omitempty"`
}

type PipelineIDValues struct {
	PipelineID []uint64 `json:"pipelineID"`
}

type AtPlanFilterStateValues struct {
	AtPlanIDs []uint64 `json:"atPlanIDs,omitempty"`
}

// Render is empty implement.
func (f *Filter) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := f.initFromProtocol(ctx, c); err != nil {
		return err
	}

	h := gshelper.NewGSHelper(gs)
	f.State.Conditions = []filter.PropCondition{
		{
			EmptyText: cputil.I18n(ctx, "all"),
			Fixed:     true,
			Key:       "pipelineID",
			Label:     cputil.I18n(ctx, "Test Plan"),
			Options: func() (opts []filter.PropConditionOption) {
				if len(f.State.Conditions[0].Options) == 0 {
					data := h.GetSelectChartHistoryData()
					opts = append(opts, filter.PropConditionOption{
						Label: data.Name,
						Value: data.PipelineID,
					})
					h.SetSelectChartItemData(gshelper.SelectChartItemData{
						PipelineID: data.PipelineID,
					})
				} else {
					opts = f.State.Conditions[0].Options
					h.SetSelectChartItemData(gshelper.SelectChartItemData{
						PipelineID: uint64(f.State.Conditions[0].Options[0].Value.(float64)),
					})
				}
				return

			}(),
			Type: filter.PropConditionTypeSelect,
		},
	}

	if err := f.setToComponent(c); err != nil {
		return err
	}
	return nil
}

func (f *Filter) initFromProtocol(ctx context.Context, c *cptype.Component) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, f); err != nil {
		return err
	}
	return nil
}

func (f *Filter) setToComponent(c *cptype.Component) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}
	return nil
}
