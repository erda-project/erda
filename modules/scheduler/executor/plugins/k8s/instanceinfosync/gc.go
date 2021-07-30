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

package instanceinfosync

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/instanceinfo"
)

// gcDeadInstancesInDB Recover the instance of phase=Dead 15 days ago
func gcDeadInstancesInDB(dbclient *instanceinfo.Client) error {
	r := dbclient.InstanceReader()
	w := dbclient.InstanceWriter()

	instances, err := r.ByPhase(instanceinfo.InstancePhaseDead).ByFinishedTime(7).Do()
	if err != nil {
		return err
	}
	ids := []uint64{}
	for _, ins := range instances {
		ids = append(ids, ins.ID)
	}
	if err := w.Delete(ids...); err != nil {
		return err
	}
	return nil
}

// gcAliveInstancesInDB Handle non-dead instances that have not been updated in a period of time and set them as dead.
// Description of non-dead instances that have not been updated in a period of time:
// 1. There is actually no pod related to this instance in the k8s cluster, and it is dead.
// 2. The event when this instance was deleted was not received or was processed correctly
// 3. This db record will never be updated again
func gcAliveInstancesInDB(dbclient *instanceinfo.Client, secs int, clustername string) error {
	r := dbclient.InstanceReader()
	w := dbclient.InstanceWriter()

	instances, err := r.ByPhases(
		instanceinfo.InstancePhaseHealthy,
		instanceinfo.InstancePhaseRunning,
		instanceinfo.InstancePhaseUnHealthy,
	).ByUpdatedTime(secs).ByTaskID(apistructs.K8S).ByCluster(clustername).Do()
	if err != nil {
		return err
	}
	for _, ins := range instances {
		ins.Phase = instanceinfo.InstancePhaseDead
		finished := time.Now()
		ins.FinishedAt = &finished
		ins.ExitCode = 255
		ins.Message = "set Dead by gc"
		if err := w.Update(ins); err != nil {
			logrus.Errorf("failed to update instance: %v", ins)
		}
	}
	return nil
}

func gcPodsInDB(dbclient *instanceinfo.Client, secs int, clustername string) error {
	r := dbclient.PodReader()
	w := dbclient.PodWriter()

	pods, err := r.ByUpdatedTime(secs).ByCluster(clustername).Do()
	if err != nil {
		return err
	}
	podIDs := []uint64{}
	for _, pod := range pods {
		podIDs = append(podIDs, pod.ID)
	}
	return w.Delete(podIDs...)
}

// gcServicesInDB The contents of the s_service_info table will not be deleted regularly, because the contents of this table should grow very slowly
func gcServicesInDB(dbclient *instanceinfo.Client) error {
	return nil
}
