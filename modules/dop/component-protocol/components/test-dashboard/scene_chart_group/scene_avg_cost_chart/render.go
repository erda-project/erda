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

package scene_avg_cost_chart

import (
	"context"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	autotestv2 "github.com/erda-project/erda/modules/dop/services/autotest_v2"
	"sort"
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common/gshelper"
	common2 "github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/scene_chart_group/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator(common.ScenarioKeyTestDashboard, "scene_avg_cost_chart", func() servicehub.Provider {
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

	scenes := h.GetAtScene()
	sceneIDs := make([]uint64, 0, len(scenes))
	sceneMap := make(map[uint64]string, 0)
	for _, v := range scenes {
		sceneMap[v.ID] = v.Name
		sceneIDs = append(sceneIDs, v.ID)
	}

	atSvc := ctx.Value(types.AutoTestPlanService).(*autotestv2.Service)
	costTimeAvg, err := atSvc.ExecHistorySceneAvgCostTime(sceneIDs...)
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
		values = append(values, int64(v.Avg))
		categories = append(categories, sceneMap[v.SceneID])
	}

	c.Props = common2.NewBarProps(values, categories, "场景 - 按执行平均耗时分布 Top500")
	return nil
}

func getCostTime(costTimeSec float64) string {
	if costTimeSec < 0 {
		return "0"
	}
	return time.Unix(int64(costTimeSec), 0).In(time.UTC).Format("15:04:05")

}
