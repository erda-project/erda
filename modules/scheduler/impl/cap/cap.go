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
