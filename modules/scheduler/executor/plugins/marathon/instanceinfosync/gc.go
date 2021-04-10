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

// gcAliveInstancesInDB Process non-Dead instances that have not been updated in a period of time.
// Strategy:
// 1. Find all non-Dead instances that have not been updated in SECS (parameters)
// 2. Check whether there are other instances with updated status (within 1 hour) on the corresponding host of each instance found above
// 3. If there is, set this instance to Dead. If not, ignore
func gcAliveInstancesInDB(dbclient *instanceinfo.Client, secs int) error {
	r := dbclient.InstanceReader()
	w := dbclient.InstanceWriter()

	instances, err := r.ByPhases(
		instanceinfo.InstancePhaseHealthy,
		instanceinfo.InstancePhaseRunning,
		instanceinfo.InstancePhaseUnHealthy,
	).ByUpdatedTime(secs).ByNotTaskID(apistructs.K8S).Do()
	if err != nil {
		return err
	}
	for _, ins := range instances {
		sameHostInstances, err := r.ByHostIP(ins.HostIP).ByNotTaskID(apistructs.K8S).ByUpdatedTime(3600).Do()
		if err != nil {
			return err
		}
		if len(sameHostInstances) > 0 {
			ins.Phase = instanceinfo.InstancePhaseDead
			finished := time.Now()
			ins.FinishedAt = &finished
			ins.ExitCode = 255
			ins.Message = "set Dead by gc"
			if err := w.Update(ins); err != nil {
				logrus.Errorf("failed to update instance: %v", ins)
			}
		}
	}
	return nil
}
