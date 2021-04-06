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

package reconciler

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (r *Reconciler) doCompensateIfHave(pCtx context.Context, pipelineID uint64) {

	pipelineWithTasks, err := r.dbClient.GetPipelineWithTasks(pipelineID)
	if err != nil {
		logrus.Errorf("failed to doCompensateIfHave, failed to get pipelineWithTasks, err: %v", err)
		return
	}

	if pipelineWithTasks == nil || pipelineWithTasks.Pipeline == nil || pipelineWithTasks.Pipeline.CronID == nil {
		return
	}

	logrus.Infof("[doCompensateIfHave] get cronID from etcd. if have compensate")
	//监听是否在执行的时候阻塞了补偿，阻塞了就立马补偿下, 获取etcd中当前pipeline的cron的id，然后立马删除对应的etcd的值
	if err := r.js.Get(pCtx, fmt.Sprint(EtcdNeedCompensatePrefix, *pipelineWithTasks.Pipeline.CronID), nil); err == nil {
		//移除cronId
		if err := r.js.Remove(pCtx, fmt.Sprint(EtcdNeedCompensatePrefix, *pipelineWithTasks.Pipeline.CronID), nil); err != nil {
			logrus.Infof("[doCompensateIfHave] can not delete etcd key, key: %s, cronId: %d, error : %v",
				fmt.Sprint(EtcdNeedCompensatePrefix, *pipelineWithTasks.Pipeline.CronID), *pipelineWithTasks.Pipeline.CronID, err)
		}

		logrus.Infof("[doCompensateIfHave] ready to Compensate, cronId: %d", *pipelineWithTasks.Pipeline.CronID)

		//执行补偿的操作
		if err := r.pipelineSvcFunc.CronNotExecuteCompensate(*pipelineWithTasks.Pipeline.CronID); err != nil {
			logrus.Infof("[doCompensateIfHave] to Compensate error, cronId: %d, error : %v",
				*pipelineWithTasks.Pipeline.CronID, err)
		}

	} else {

		logrus.Infof("[doCompensateIfHave]: can not get cronId %d err: %v", *pipelineWithTasks.Pipeline.CronID, err)
	}
}
