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
