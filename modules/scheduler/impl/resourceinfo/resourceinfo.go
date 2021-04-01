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
