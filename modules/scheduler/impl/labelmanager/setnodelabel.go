package labelmanager

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/impl/cluster/clusterutil"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	NotFoundExecutor = errors.New("Not found Executor")
)

func (s *LabelManagerImpl) SetNodeLabel(cluster Cluster, hosts []string, tags map[string]string) error {
	executorType := clusterutil.ServiceKindMarathon
	switch strutil.ToLower(cluster.ClusterType) {
	case "dcos":
		executorType = clusterutil.ServiceKindMarathon
	case apistructs.K8S, "kubernetes":
		// executorType = clusterutil.ServiceKindK8S
		// TODO: 现在即使是 k8s， executorname也是 MARATHONFORXXX
		executorType = clusterutil.ServiceKindMarathon
	case "edas":
		executorType = clusterutil.EdasKindInK8s
	}
	executorName := clusterutil.GenerateExecutorByCluster(cluster.ClusterName, executorType)
	executor, err := executor.GetManager().Get(executortypes.Name(executorName))
	if err != nil {
		return errors.Wrap(NotFoundExecutor, executorName)
	}
	if err := executor.SetNodeLabels(executortypes.NodeLabelSetting{cluster.SoldierURL}, hosts, tags); err != nil {
		return err
	}
	return nil
}
