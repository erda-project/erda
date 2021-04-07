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

// gcDeadInstancesInDB 回收 15 天前的 phase=Dead 的实例
func gcDeadInstancesInDB(dbclient *instanceinfo.Client) error {
	r := dbclient.InstanceReader()
	w := dbclient.InstanceWriter()

	instances, err := r.ByPhase(instanceinfo.InstancePhaseDead).ByFinishedTime(15).Do()
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

// gcAliveInstancesInDB 处理一段时间内没有更新过的非 dead 实例, 将它们置为 dead.
// 一段时间内都没有更新过的非 dead 实例说明:
// 1. k8s集群中实际已经没有这个实例相关pod了, 已经 dead 了
// 2. 这个实例被删除的时候的事件没有收到或者被正确处理
// 3. 这条 db 记录再也不会被更新了
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

// gcServicesInDB 不会定期删除 s_service_info 表里内容, 因为这个表内容应该是增长非常慢的
func gcServicesInDB(dbclient *instanceinfo.Client) error {
	return nil
}
