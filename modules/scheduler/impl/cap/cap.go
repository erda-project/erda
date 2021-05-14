// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cap

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/impl/cluster/clusterutil"
)

type CapImpl struct{}

type Cap interface {
	CapacityInfo(clustername string) apistructs.CapacityInfoData
}

func NewCapImpl() Cap {
	return &CapImpl{}
}

func (cap *CapImpl) CapacityInfo(clustername string) apistructs.CapacityInfoData {
	executorname := clusterutil.GenerateExecutorByCluster(clustername, clusterutil.ServiceKindMarathon)
	extor, err := executor.GetManager().Get(executortypes.Name(executorname))
	if err != nil {
		return apistructs.CapacityInfoData{}
	}
	return extor.CapacityInfo()
}
