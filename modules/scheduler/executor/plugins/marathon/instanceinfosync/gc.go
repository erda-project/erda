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
