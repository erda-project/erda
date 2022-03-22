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

package reconciler

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (r *Reconciler) doCronCompensate(pCtx context.Context, pipelineID uint64) {

	pipelineWithTasks, err := r.dbClient.GetPipelineWithTasks(pipelineID)
	if err != nil {
		logrus.Errorf("failed to doCronCompensate, failed to get pipelineWithTasks, err: %v", err)
		return
	}

	if pipelineWithTasks == nil || pipelineWithTasks.Pipeline == nil || pipelineWithTasks.Pipeline.CronID == nil {
		return
	}

	logrus.Infof("[doCronCompensate] get cronID from etcd. if have compensate")

	// Monitor whether the compensation is blocked during execution, and immediately compensate if it is blocked,
	// obtain the id of the cron of the current pipeline in etcd, and then immediately delete the corresponding etcd value
	notFound, err := r.js.Notfound(pCtx, fmt.Sprint(EtcdNeedCompensatePrefix, *pipelineWithTasks.Pipeline.CronID))
	if err != nil {
		logrus.Infof("[doCronCompensate]: can not get cronID: %d, err: %v", *pipelineWithTasks.Pipeline.CronID, err)
		return
	}

	if !notFound {
		logrus.Infof("[doCronCompensate] ready to Compensate, cronID: %d", *pipelineWithTasks.Pipeline.CronID)

		// perform the compensation operation
		if err := r.pipelineSvcFunc.CronNotExecuteCompensate(*pipelineWithTasks.Pipeline.CronID); err != nil {
			logrus.Infof("[doCronCompensate] to Compensate error, cronID: %d, err : %v",
				*pipelineWithTasks.Pipeline.CronID, err)
		}

		// remove cronID
		if err := r.js.Remove(pCtx, fmt.Sprint(EtcdNeedCompensatePrefix, *pipelineWithTasks.Pipeline.CronID), nil); err != nil {
			logrus.Infof("[doCronCompensate] can not delete etcd key, key: %s, cronID: %d, err : %v",
				fmt.Sprint(EtcdNeedCompensatePrefix, *pipelineWithTasks.Pipeline.CronID), *pipelineWithTasks.Pipeline.CronID, err)
		}
	}
}
