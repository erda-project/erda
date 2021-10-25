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

package at_api_avg_cost_chart

import (
	"context"
	"sort"
	"strconv"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	autotestv2 "github.com/erda-project/erda/modules/dop/services/autotest_v2"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator(common.ScenarioKeyTestDashboard, "at_api_avg_cost_chart", func() servicehub.Provider {
		return &Chart{}
	})
}

type Chart struct {
	base.DefaultProvider

	Values     []string `json:"values"`
	Categories []string `json:"categories"`
}

func (f *Chart) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	h := gshelper.NewGSHelper(gs)

	sceneSteps := h.GetGlobalAtSceneStep()
	sceneStepIDs := make([]uint64, 0, len(sceneSteps))
	sceneStepMap := make(map[uint64]string, 0)
	for _, v := range sceneSteps {
		sceneStepMap[v.ID] = v.Name
		sceneStepIDs = append(sceneStepIDs, v.ID)
	}

	timeFilter := h.GetAtSceneAndApiTimeFilter()
	atSvc := ctx.Value(types.AutoTestPlanService).(*autotestv2.Service)
	projectID, _ := strconv.ParseUint(cputil.GetInParamByKey(ctx, "projectID").(string), 10, 64)
	costTimeAvg, err := atSvc.ExecHistoryApiAvgCostTime(apistructs.StatisticsExecHistoryRequest{
		TimeStart:    timeFilter.TimeStart,
		TimeEnd:      timeFilter.TimeEnd,
		IterationIDs: h.GetGlobalSelectedIterationIDs(),
		PlanIDs:      h.GetGlobalAutoTestPlanIDs(),
		SceneSetIDs:  nil,
		SceneIDs:     nil,
		StepIDs:      nil,
		ProjectID:    projectID,
	})
	if err != nil {
		return err
	}
	sort.Slice(costTimeAvg, func(i, j int) bool {
		return costTimeAvg[i].Avg > costTimeAvg[j].Avg
	})

	var (
		values     []int64
		categories []string
	)
	for _, v := range costTimeAvg {
		if _, ok := sceneStepMap[v.StepID]; !ok {
			continue
		}
		values = append(values, int64(v.Avg))
		categories = append(categories, sceneStepMap[v.StepID])
	}

	c.Props = common.NewBarProps(values, categories, cputil.I18n(ctx, "api-avg-cost"))
	return nil
}
