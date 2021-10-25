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

package at_block_header_filter

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
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	autotestv2 "github.com/erda-project/erda/modules/dop/services/autotest_v2"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

func init() {
	base.InitProviderWithCreator(common.ScenarioKeyTestDashboard, "at_block_header_filter", func() servicehub.Provider {
		return &Filter{}
	})
}

type Filter struct {
	base.DefaultProvider

	State      State `json:"state,omitempty"`
	atTestPlan *autotestv2.Service
}

type State struct {
	Conditions []filter.PropCondition  `json:"conditions,omitempty"`
	Values     AtPlanFilterStateValues `json:"values,omitempty"`
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
	globalAtPlans := h.GetGlobalAutoTestPlanList()
	//selectAtPlans := h.GetAtBlockFilterTestPlanList()
	selectedItrsByID := h.GetGlobalSelectedIterationsByID()
	f.State.Conditions = []filter.PropCondition{
		{
			EmptyText: cputil.I18n(ctx, "all"),
			Fixed:     true,
			Key:       "atPlanIDs",
			Label:     cputil.I18n(ctx, "Test Plan"),
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
	}
	// put selected values into global state
	h.SetAtBlockFilterTestPlanList(func() (selectedAtPlans []apistructs.TestPlanV2) {
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
	if err := f.SetBlockAtSceneAndStep(h); err != nil {
		return err
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

	f.atTestPlan = ctx.Value(types.AutoTestPlanService).(*autotestv2.Service)
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

func (f *Filter) SetBlockAtSceneAndStep(h *gshelper.GSHelper) error {
	steps, err := f.atTestPlan.ListStepByPlanID(func() []uint64 {
		selectPlans := h.GetAtBlockFilterTestPlanList()
		selectPlanIDs := make([]uint64, 0, len(selectPlans))
		for _, v := range selectPlans {
			selectPlanIDs = append(selectPlanIDs, v.ID)
		}
		return selectPlanIDs
	}()...)
	if err != nil {
		return err
	}
	scenes, err := f.atTestPlan.ListSceneBySceneSetID(func() []uint64 {
		setIDs := make([]uint64, 0, len(steps))
		for _, v := range steps {
			setIDs = append(setIDs, v.SceneSetID)
		}
		return setIDs
	}()...)
	if err != nil {
		return err
	}

	sceneSteps, err := f.atTestPlan.ListAutoTestSceneSteps(func() []uint64 {
		sceneIDs := make([]uint64, 0, len(scenes))
		for _, v := range scenes {
			sceneIDs = append(sceneIDs, v.ID)
		}
		return sceneIDs
	}())
	if err != nil {
		return err
	}

	h.SetBlockAtStep(steps)
	h.SetBlockAtScene(scenes)
	h.SetBlockAtSceneStep(sceneSteps)
	return nil
}
