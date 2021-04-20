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
		// TODO: Now even for k8s, executorname is MARATHONFORXXX
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
