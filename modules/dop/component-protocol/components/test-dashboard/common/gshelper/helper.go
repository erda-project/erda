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

package gshelper

import (
	"github.com/mitchellh/mapstructure"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
)

type GSHelper struct {
	gs *cptype.GlobalStateData
}

func NewGSHelper(gs *cptype.GlobalStateData) *GSHelper {
	return &GSHelper{gs: gs}
}

func assign(src, dst interface{}) error {
	if src == nil || dst == nil {
		return nil
	}
	return mapstructure.Decode(src, dst)
}

func (h *GSHelper) SetGlobalManualTestPlanList(l []apistructs.TestPlan) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalManualTestPlanList"] = l
}

func (h *GSHelper) GetGlobalManualTestPlanList() []apistructs.TestPlan {
	if h.gs == nil {
		return nil
	}
	res := make([]apistructs.TestPlan, 0)
	_ = assign((*h.gs)["GlobalManualTestPlanList"], &res)
	return res
}

func (h *GSHelper) SetMtBlockFilterTestPlanList(l []apistructs.TestPlan) {
	if h.gs == nil {
		return
	}
	(*h.gs)["MtBlockFilterTestPlanList"] = l
}

func (h *GSHelper) GetMtBlockFilterTestPlanList() []apistructs.TestPlan {
	if h.gs == nil {
		return nil
	}
	res := make([]apistructs.TestPlan, 0)
	_ = assign((*h.gs)["MtBlockFilterTestPlanList"], &res)
	return res
}

func (h *GSHelper) SetGlobalSelectedIterationIDs(ids []uint64) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalSelectedIterationIDs"] = ids
}

func (h *GSHelper) GetGlobalSelectedIterationIDs() []uint64 {
	if h.gs == nil {
		return nil
	}
	res := make([]uint64, 0)
	_ = assign((*h.gs)["GlobalSelectedIterationIDs"], &res)
	return res
}

func (h *GSHelper) SetGlobalSelectedIterationsByID(l map[uint64]dao.Iteration) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalSelectedIterationsByID"] = l
}

func (h *GSHelper) GetGlobalSelectedIterationsByID() map[uint64]dao.Iteration {
	if h.gs == nil {
		return nil
	}
	res := make(map[uint64]dao.Iteration, 0)
	_ = assign((*h.gs)["GlobalSelectedIterationsByID"], &res)
	return res
}

func (h *GSHelper) SetMtPlanChartFilterTestPlanList(l []apistructs.TestPlan) {
	if h.gs == nil {
		return
	}
	(*h.gs)["MtPlanChartFilterTestPlanList"] = l
}

func (h *GSHelper) GetMtPlanChartFilterTestPlanList() []apistructs.TestPlan {
	if h.gs == nil {
		return nil
	}
	res := make([]apistructs.TestPlan, 0)
	_ = assign((*h.gs)["MtPlanChartFilterTestPlanList"], &res)
	return res
}

func (h *GSHelper) SetMtPlanChartFilterStatusList(l []apistructs.TestCaseExecStatus) {
	if h.gs == nil {
		return
	}
	(*h.gs)["MtPlanChartFilterStatusList"] = l
}

func (h *GSHelper) GetMtPlanChartFilterStatusList() []apistructs.TestCaseExecStatus {
	if h.gs == nil {
		return nil
	}
	res := make([]apistructs.TestCaseExecStatus, 0)
	_ = assign((*h.gs)["MtPlanChartFilterStatusList"], &res)
	return res
}

func (h *GSHelper) SetGlobalAutoTestPlanList(l []*apistructs.TestPlanV2) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalAutoTestPlanList"] = l
}

func (h *GSHelper) GetGlobalAutoTestPlanList() []apistructs.TestPlanV2 {
	if h.gs == nil {
		return nil
	}
	res := make([]apistructs.TestPlanV2, 0)
	_ = assign((*h.gs)["GlobalAutoTestPlanList"], &res)
	return res
}

func (h *GSHelper) SetGlobalAutoTestPlanIDs(l []uint64) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalAutoTestPlanIDs"] = l
}

func (h *GSHelper) GetGlobalAutoTestPlanIDs() []uint64 {
	if h.gs == nil {
		return nil
	}
	res := make([]uint64, 0)
	_ = assign((*h.gs)["GlobalAutoTestPlanIDs"], &res)
	return res
}

func (h *GSHelper) SetAtBlockFilterTestPlanList(l []apistructs.TestPlanV2) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalAtBlockFilterTestPlanList"] = l
}

func (h *GSHelper) GetAtBlockFilterTestPlanList() []apistructs.TestPlanV2 {
	if h.gs == nil {
		return nil
	}
	res := make([]apistructs.TestPlanV2, 0)
	_ = assign((*h.gs)["GlobalAtBlockFilterTestPlanList"], &res)
	return res
}

func (h *GSHelper) SetAtStep(l []apistructs.TestPlanV2Step) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalAtStep"] = l
}

func (h *GSHelper) GetAtStep() []apistructs.TestPlanV2Step {
	if h.gs == nil {
		return nil
	}
	res := make([]apistructs.TestPlanV2Step, 0)
	_ = assign((*h.gs)["GlobalAtStep"], &res)
	return res
}

func (h *GSHelper) SetAtScene(l []apistructs.AutoTestScene) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalAtScene"] = l
}

func (h *GSHelper) GetAtScene() []apistructs.AutoTestScene {
	if h.gs == nil {
		return nil
	}
	res := make([]apistructs.AutoTestScene, 0)
	_ = assign((*h.gs)["GlobalAtScene"], &res)
	return res
}

func (h *GSHelper) SetAtSceneStep(l []apistructs.AutoTestSceneStep) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalAtSceneStep"] = l
}

func (h *GSHelper) GetAtSceneStep() []apistructs.AutoTestSceneStep {
	if h.gs == nil {
		return nil
	}
	res := make([]apistructs.AutoTestSceneStep, 0)
	_ = assign((*h.gs)["GlobalAtSceneStep"], &res)
	return res
}

type AtSceneAndApiTimeFilter struct {
	TimeStart string `json:"timeStart"`
	TimeEnd   string `json:"timeEnd"`
}

func (h *GSHelper) SetAtSceneAndApiTimeFilter(t AtSceneAndApiTimeFilter) {
	if h.gs == nil {
		return
	}
	(*h.gs)["AtSceneAndApiTimeFilter"] = t
}

func (h *GSHelper) GetAtSceneAndApiTimeFilter() AtSceneAndApiTimeFilter {
	if h.gs == nil {
		return AtSceneAndApiTimeFilter{}
	}
	return (*h.gs)["AtSceneAndApiTimeFilter"].(AtSceneAndApiTimeFilter)
}

func (h *GSHelper) SetAtCaseRateTrendingTimeFilter(t AtSceneAndApiTimeFilter) {
	if h.gs == nil {
		return
	}
	(*h.gs)["AtCaseRateTrendingTimeFilter"] = t
}

func (h *GSHelper) GetAtCaseRateTrendingTimeFilter() AtSceneAndApiTimeFilter {
	if h.gs == nil {
		return AtSceneAndApiTimeFilter{}
	}
	return (*h.gs)["AtCaseRateTrendingTimeFilter"].(AtSceneAndApiTimeFilter)
}

type SelectChartItemData struct {
	PlanID     uint64 `json:"planId"`
	Name       string `json:"name"`
	PipelineID uint64 `json:"pipelineID"`
}

func (h *GSHelper) SetSelectChartItemData(t SelectChartItemData) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalSelectChartItemData"] = t
}

func (h *GSHelper) GetSelectChartHistoryData() SelectChartItemData {
	if h.gs == nil {
		return SelectChartItemData{}
	}
	return (*h.gs)["GlobalSelectChartItemData"].(SelectChartItemData)
}
