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
