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

package resourceinfo

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/impl/cluster/clusterutil"
)

type ResourceInfo interface {
	Info(clusterName string, brief bool) (apistructs.ClusterResourceInfoData, error)
}

type ResourceInfoImpl struct {
}

func NewResourceInfoImpl() ResourceInfo {
	return &ResourceInfoImpl{}
}

func (r *ResourceInfoImpl) Info(clusterName string, brief bool) (apistructs.ClusterResourceInfoData, error) {
	executorName := clusterutil.GenerateExecutorByClusterName(clusterName)
	executor, err := executor.GetManager().Get(executortypes.Name(executorName))
	if err != nil {
		return apistructs.ClusterResourceInfoData{}, fmt.Errorf("not found executor: %s", executorName)
	}
	return executor.ResourceInfo(brief)
}
