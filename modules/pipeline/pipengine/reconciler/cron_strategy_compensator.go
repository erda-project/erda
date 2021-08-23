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
	//监听是否在执行的时候阻塞了补偿，阻塞了就立马补偿下, 获取etcd中当前pipeline的cron的id，然后立马删除对应的etcd的值
	if err := r.js.Get(pCtx, fmt.Sprint(EtcdNeedCompensatePrefix, *pipelineWithTasks.Pipeline.CronID), nil); err == nil {
		//移除cronId
		if err := r.js.Remove(pCtx, fmt.Sprint(EtcdNeedCompensatePrefix, *pipelineWithTasks.Pipeline.CronID), nil); err != nil {
			logrus.Infof("[doCronCompensate] can not delete etcd key, key: %s, cronId: %d, error : %v",
				fmt.Sprint(EtcdNeedCompensatePrefix, *pipelineWithTasks.Pipeline.CronID), *pipelineWithTasks.Pipeline.CronID, err)
		}

		logrus.Infof("[doCronCompensate] ready to Compensate, cronId: %d", *pipelineWithTasks.Pipeline.CronID)

		//执行补偿的操作
		if err := r.pipelineSvcFunc.CronNotExecuteCompensate(*pipelineWithTasks.Pipeline.CronID); err != nil {
			logrus.Infof("[doCronCompensate] to Compensate error, cronId: %d, error : %v",
				*pipelineWithTasks.Pipeline.CronID, err)
		}

	} else {

		logrus.Infof("[doCronCompensate]: can not get cronId %d err: %v", *pipelineWithTasks.Pipeline.CronID, err)
	}
}
