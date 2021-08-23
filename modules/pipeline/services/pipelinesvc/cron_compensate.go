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

package pipelinesvc

import (
	"context"
	"fmt"
	"strconv"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/dlock"
	"github.com/erda-project/erda/pkg/limit_sync_group"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	cronCompensateDLockKey = "/devops/pipeline/crond/compensator"
	waitTimeIfLostDLock    = time.Minute
	waitTimeIfQueryDBError = time.Minute
	indexFirst             = 0 //数组的下标0
)

var compensateLog = logrus.WithField("type", "cron compensator")

func getCronCompensateInterval(interval int64) time.Duration {
	return time.Duration(interval) * time.Minute
}

func getCronInterruptCompensateInterval(interval int64) time.Duration {
	return time.Duration(interval) * time.Minute * 2
}

func (s *PipelineSvc) ContinueCompensate() {
	// 获取分布式锁成功才能执行中断补偿
	// 若分布式锁丢失，停止补偿，并尝试重新获取分布式锁
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//lock, err := s.getCompensateLock(cancel)
	//if err != nil {
	//	return
	//}
	//
	//defer func() {
	//	if lock != nil {
	//		lock.UnlockAndClose()
	//	}
	//}()

	compensateLog.Info("cron compensator: start")
	//执行策略补偿定时任务
	ticker := time.NewTicker(getCronCompensateInterval(conf.CronCompensateTimeMinute()))
	//中断补偿的定时任务
	interruptTicker := time.NewTicker(getCronInterruptCompensateInterval(conf.CronCompensateTimeMinute()))

	//项目启动先执行一次中断补偿
	s.traverseDoCompensate(s.doInterruptCompensate, true)

	for {

		select {
		case <-ctx.Done():
			// stop
			compensateLog.Info("stop cron compensate, received cancel signal from channel")
			ticker.Stop()
			interruptTicker.Stop()
			return
		case <-interruptTicker.C:
			//这里中断补偿为啥用同步，因为这里会新增数据，目前还不确认是否有幂等，所以先同步
			s.traverseDoCompensate(s.doInterruptCompensate, true)
		case <-ticker.C:
			//因为未执行补偿(策略)，用来执行流水线的，内部是幂等的，也就是同一时间只有一个pipeline在执行
			s.traverseDoCompensate(s.doStrategyCompensate, false)
		}
	}
}

//获取etcd到全局锁，代表补偿只有一个实例能执行
func (s *PipelineSvc) getCompensateLock(cancel func()) (*dlock.DLock, error) {
	lock, err := dlock.New(
		cronCompensateDLockKey,
		func() {
			compensateLog.Error("[alert] dlock lost, stop current compensate")
			cancel()
			time.Sleep(waitTimeIfLostDLock)
			compensateLog.Warn("try to continue compensate again")
			go s.ContinueCompensate()
		},
		dlock.WithTTL(30),
	)
	if err != nil {
		compensateLog.Errorf("[alert] failed to get dlock, err: %v", err)
		time.Sleep(waitTimeIfLostDLock)
		go s.ContinueCompensate()
		return nil, err
	}
	if err := lock.Lock(context.Background()); err != nil {
		compensateLog.Errorf("[alert] failed to lock dlock, err: %v", err)
		time.Sleep(waitTimeIfLostDLock)
		go s.ContinueCompensate()
		return nil, err
	}

	return lock, nil
}

func (s *PipelineSvc) doInterruptCompensate(pc spec.PipelineCron) {
	// 中断补偿
	err := s.cronInterruptCompensate(pc)
	if err != nil {
		compensateLog.WithField("cronID", pc.ID).Errorf("failed to do interrupt-compensate, cronID: %d, err: %v", pc.ID, err)
	}
}

func (s *PipelineSvc) doStrategyCompensate(pc spec.PipelineCron) {
	//开始为每个 enable 且开启补偿的 cron 执行 未执行补偿 检查
	err := s.cronNotExecuteCompensate(pc)
	if err != nil {
		compensateLog.WithField("cronID", pc.ID).Errorf("failed to do notexecute-compensate, cronID: %d, err: %v", pc.ID, err)
	}
}

func (s *PipelineSvc) traverseDoCompensate(doCompensate func(cron spec.PipelineCron), sync bool) {

	if doCompensate == nil {
		return
	}

	// get all enabled crons
	enabledCrons, err := s.dbClient.ListPipelineCrons(&[]bool{true}[0])
	if err != nil {
		compensateLog.Errorf("failed to list enabled pipeline crons from db, try again later, err: %v", err)
		time.Sleep(waitTimeIfQueryDBError)
		return
	}

	group := limit_sync_group.NewSemaphore(int(conf.CronCompensateConcurrentNumber()))
	for i := range enabledCrons {
		if sync {
			doCompensate(enabledCrons[i])
		} else {
			group.Add(1)
			go func(pc spec.PipelineCron) {
				defer group.Done()
				doCompensate(pc)
			}(enabledCrons[i])
		}
	}
	group.Wait()
}

// cronInterruptCompensate 定时 中断补偿
func (s *PipelineSvc) cronInterruptCompensate(pc spec.PipelineCron) error {

	// 计算中断补偿开始时间
	beforeCompensateFromTime := getCompensateFromTime(pc)

	// current time minus a certain time，prevent conflicts with cron create pipeline
	now := time.Unix(time.Now().Unix(), 0)
	var thisCompensateFromTime = now.Add(time.Second * -time.Duration(conf.CronFailureCreateIntervalCompensateTimeSecond()))

	// 用 cron expr 计算出从起始补偿点开始的所有需要触发时间
	needTriggerTimes, err := pipelineyml.ListNextCronTime(pc.CronExpr,
		pipelineyml.WithCronStartEndTime(&beforeCompensateFromTime, &thisCompensateFromTime),
		pipelineyml.WithListNextScheduleCount(100),
	)
	if err != nil {
		return errors.Errorf("[alert] failed to list next crontimes, cronID: %d, err: %v", pc.ID, err)
	}

	if len(needTriggerTimes) > 0 {
		// 根据 source + ymlName + timeCreated 搜索已经创建的流水线记录
		existPipelines, _, _, _, err := s.dbClient.PageListPipelines(apistructs.PipelinePageListRequest{
			Sources:          []apistructs.PipelineSource{pc.PipelineSource},
			YmlNames:         []string{pc.PipelineYmlName},
			TriggerModes:     []apistructs.PipelineTriggerMode{apistructs.PipelineTriggerModeCron},
			StartTimeCreated: beforeCompensateFromTime,
			EndTimeCreated:   thisCompensateFromTime,
			PageNum:          1,
			PageSize:         100,
			LargePageSize:    true,
		})
		if err != nil {
			return errors.Errorf("[alert] failed to list existPipelines, cronID: %d, err: %v", pc.ID, err)
		}

		// 转换为 map 用于查询
		existPipelinesMap := make(map[time.Time]spec.Pipeline, len(existPipelines))
		for _, p := range existPipelines {
			existPipelinesMap[getTriggeredTime(p)] = p
		}

		// 遍历 needTriggerTimes，若没创建，则需要中断补偿创建
		for _, ntt := range needTriggerTimes {
			p, ok := existPipelinesMap[ntt]
			if ok {
				compensateLog.Infof("no need do interrupt-compensate, cronID: %d, triggerTime: %v, exist pipelineID: %d", pc.ID, ntt, p.ID)
				continue
			}
			compensateLog.Infof("need do interrupt-compensate, cronID: %d, triggerTime: %v", pc.ID, ntt)
			// create
			created, err := s.createCronCompensatePipeline(pc, ntt)
			if err != nil {
				compensateLog.Errorf("failed to do interrupt-compensate, cronID: %d, triggerTime: %v, err: %v", pc.ID, ntt, err)
				continue
			}
			compensateLog.Infof("success to do interrupt-compensate, cronID: %d, triggerTime: %v, createdPipelineID: %d", pc.ID, ntt, created.ID)
		}
	}

	// 中断补偿完毕，需要更新 cron 的 thisCompensateFromTime 字段
	pc.Extra.LastCompensateAt = &thisCompensateFromTime
	// 若 compensator 为空，说明是老的 cron，自动使用默认配置
	if pc.Extra.Compensator == nil {
		pc.Extra.Compensator = &apistructs.CronCompensator{
			Enable:               pipelineyml.DefaultCronCompensator.Enable,
			LatestFirst:          pipelineyml.DefaultCronCompensator.LatestFirst,
			StopIfLatterExecuted: pipelineyml.DefaultCronCompensator.StopIfLatterExecuted,
		}
	}
	if err := s.dbClient.UpdatePipelineCron(pc.ID, &pc); err != nil {
		return errors.Errorf("failed to update pipelineCron for lastCompensateAt field, err: %v", err)
	}

	return nil
}

func (s *PipelineSvc) CronNotExecuteCompensateById(id uint64) error {
	cron, err := s.dbClient.GetPipelineCron(id)
	if err != nil {
		return err
	}

	return s.cronNotExecuteCompensate(cron)
}

// cronNotExecuteCompensate 定时 未执行补偿
// 只执行一天内
func (s *PipelineSvc) cronNotExecuteCompensate(pc spec.PipelineCron) error {

	// 未启用 notexecute compensate，退出
	if pc.Enable == nil || *pc.Enable == false || pc.Extra.Compensator == nil || pc.Extra.Compensator.Enable == false {
		return nil
	}

	now := time.Unix(time.Now().Unix(), 0)
	oneDayBeforeNow := now.AddDate(0, 0, -1)

	// 获取待执行列表
	// 根据 source + ymlName + id 搜索已经创建的流水线记录
	//这里为什么按id拿10个，因为在表达式粒度很小的情况下，会丢失很多数据，然而中断补偿创建的id顺序和执行顺序不同
	//所以中和先用id正序或者倒序拿10条，后面doCronCompensate再根据10条的具体执行时间挑最适合时间的进行执行
	//时间粒度很大的情况下，本质来说拿一条就Ok了
	request := apistructs.PipelinePageListRequest{
		Sources:          []apistructs.PipelineSource{pc.PipelineSource},
		YmlNames:         []string{pc.PipelineYmlName},
		Statuses:         []string{apistructs.PipelineStatusAnalyzed.String()},
		TriggerModes:     []apistructs.PipelineTriggerMode{apistructs.PipelineTriggerModeCron},
		StartTimeCreated: oneDayBeforeNow,
		EndTimeCreated:   now,
		PageNum:          1,
		PageSize:         10,
		LargePageSize:    true,
	}

	if (*pc.Extra.Compensator).LatestFirst {
		request.DescCols = []string{apistructs.PipelinePageListRequestIdColumn}
	} else {
		request.AscCols = []string{apistructs.PipelinePageListRequestIdColumn}
	}

	existPipelines, _, _, _, err := s.dbClient.PageListPipelines(request)
	if err != nil {
		return errors.Errorf("failed to list notexecute pipelines, cronID: %d, err: %v", pc.ID, err)
	}

	return s.doCronCompensate(*pc.Extra.Compensator, existPipelines, pc)
}

func (s *PipelineSvc) doCronCompensate(compensator apistructs.CronCompensator, notRunPipelines []spec.Pipeline, pipelineCron spec.PipelineCron) error {
	var order string

	if len(notRunPipelines) <= 0 {
		return nil
	}

	//根据策略从排好序的未执行中挑选出最适合时间点
	if compensator.LatestFirst {
		order = "DESC"
	} else {
		order = "ASC"
	}
	//doCronCompensate再根据10条的具体执行时间挑最适合时间的进行执行
	firstOrLastPipeline := orderByCronTriggerTime(notRunPipelines, order)[indexFirst]

	//根据策略判定假如是最后一个pipeline，当是StopIfLatterExecuted的策略的时候，应该和最新的suucess状态的pipeline进行一个时间对比，只有id 大于 成功的id才能执行
	if compensator.LatestFirst && compensator.StopIfLatterExecuted {
		// 获取执行成功的pipeline
		runSuccessPipeline, _, _, _, err := s.dbClient.PageListPipelines(apistructs.PipelinePageListRequest{
			Sources:  []apistructs.PipelineSource{pipelineCron.PipelineSource},
			YmlNames: []string{pipelineCron.PipelineYmlName},
			Statuses: []string{apistructs.PipelineStatusSuccess.String()},
			PageNum:  1,
			PageSize: 1,
			DescCols: []string{apistructs.PipelinePageListRequestIdColumn},
		})

		if err != nil {
			compensateLog.Infof("latestFirst=true, stopIfLatterExecuted=true, get PipelineStatusSuccess pipeline error, cronID: %d", pipelineCron.ID)
		}

		if len(runSuccessPipeline) <= 0 {
			return nil
		}

		//最新的成功的id 大于 补偿的 id, 就不进行补偿
		lastSuccessPipeline := runSuccessPipeline[indexFirst]
		if lastSuccessPipeline.ID > firstOrLastPipeline.ID {
			return nil
		}
	}

	_, err := s.RunPipeline(&apistructs.PipelineRunRequest{
		PipelineID:   firstOrLastPipeline.ID,
		IdentityInfo: apistructs.IdentityInfo{InternalClient: firstOrLastPipeline.Extra.InternalClient},
	})

	//执行成功打印一行记录
	if err == nil {
		compensateLog.Infof("[doCronCompensate] Compensate success, pipelineId %d", firstOrLastPipeline.ID)
		return nil
	}

	//补偿出现冲突或者内部出现错误，应该通知监听下次应该直接执行这个补偿的调度，
	//对应的cron表达式间隔和执行时间加起来大于未执行补偿的调度时间，整个调度就和配置的策略一样进行执行，小于的话，就会出现竞争执行
	compensateLog.Infof("[doCronCompensate] run Compensate err, put cronId into etcd wait callback: cronId %d", pipelineCron.ID)
	//创建etcd租约
	lease := v3.NewLease(s.etcdctl.GetClient())
	if grant, err := lease.Grant(context.Background(), conf.CronCompensateTimeMinute()*60); err == nil {
		//将cronid设置到key下，等待
		if _, err := s.js.PutWithOption(context.Background(),
			fmt.Sprint(reconciler.EtcdNeedCompensatePrefix, pipelineCron.ID),
			nil, []interface{}{v3.WithLease(grant.ID)}); err != nil {
			// 写入etcd失败，这次补偿失败，等待下次补偿
			logrus.Errorf("[alert] failed to write cronId to etcd: cronId %d, err: %v", pipelineCron.ID, err)
			return err
		}
	} else {
		logrus.Errorf("[alert] failed to create etcd lease : cronId %d, err: %v", pipelineCron.ID, err)
		return err
	}

	compensateLog.Infof("[doCronCompensate] put cronId into etcd suucess: cronId %d ", pipelineCron.ID)

	return errors.Errorf("[doCronCompensate] failed to run notexecute pipeline, cronID: %d, pipelineID: %d, err: %v", pipelineCron.ID, firstOrLastPipeline.ID, err)

}

// orderByCronTriggerTime 排序
// order: ASC/DESC
func orderByCronTriggerTime(inputs []spec.Pipeline, order string) []spec.Pipeline {
	var result []spec.Pipeline
	for i := range inputs {
		input := inputs[i]
		if input.Extra.CronTriggerTime == nil {
			continue
		}
		if len(result) == 0 {
			result = append(result, input)
			continue
		}
		if strutil.ToUpper(order) == "DESC" && (*input.Extra.CronTriggerTime).After(*result[0].Extra.CronTriggerTime) {
			result = append([]spec.Pipeline{input}, result...)
		} else {
			result = append(result, input)
		}
	}
	return result
}

// getCompensateFromTime 计算 中断补偿 起始时间点
// 若 extra.LastCompensateAt 为空，说明未补偿过，使用 cron 更新时间作为补偿起始时间
// 最多补偿一天
//
// 有一个特殊场景：
// 1. 手动修改 enable 字段为 0，重启 pipeline 使 cron 停止
// 2. 做一些需要暂时停止 cron 的操作，例如数据库迁移，集群调整等
// 3. 手动修改 enable = 1，重启 pipeline 使 cron 生效，需要做 中断补偿
// 过程中 lastCompensateAt 字段未更新，cron 被临时停止，这种情况同样使用 cron 更新时间作为 补偿起始时间
func getCompensateFromTime(pc spec.PipelineCron) (t time.Time) {
	now := time.Unix(time.Now().Unix(), 0)
	defer func() {
		if now.Sub(t) > time.Hour*24 {
			t = now.Add(-time.Hour * 24)
		}
	}()

	if pc.Extra.LastCompensateAt != nil {
		return *pc.Extra.LastCompensateAt
	}
	return pc.TimeUpdated
}

// getTriggeredTime 获取创建时间，定时创建的流水线使用 cronTriggerTime
func getTriggeredTime(p spec.Pipeline) time.Time {
	if p.Extra.CronTriggerTime != nil {
		return time.Unix((*p.Extra.CronTriggerTime).Unix(), 0)
	}
	return time.Unix(p.TimeCreated.Unix(), 0)
}

func (s *PipelineSvc) createCronCompensatePipeline(pc spec.PipelineCron, triggerTime time.Time) (*spec.Pipeline, error) {
	// cron
	if pc.Extra.NormalLabels == nil {
		pc.Extra.NormalLabels = make(map[string]string)
	}
	if pc.Extra.FilterLabels == nil {
		pc.Extra.FilterLabels = make(map[string]string)
	}
	if _, ok := pc.Extra.FilterLabels[apistructs.LabelPipelineTriggerMode]; ok {
		pc.Extra.FilterLabels[apistructs.LabelPipelineTriggerMode] = apistructs.PipelineTriggerModeCron.String()
	}
	pc.Extra.NormalLabels[apistructs.LabelPipelineTriggerMode] = apistructs.PipelineTriggerModeCron.String()
	pc.Extra.NormalLabels[apistructs.LabelPipelineType] = apistructs.PipelineTypeNormal.String()
	pc.Extra.NormalLabels[apistructs.LabelPipelineYmlSource] = apistructs.PipelineYmlSourceContent.String()
	pc.Extra.NormalLabels[apistructs.LabelPipelineCronTriggerTime] = strconv.FormatInt(triggerTime.UnixNano(), 10)
	pc.Extra.NormalLabels[apistructs.LabelPipelineCronID] = strconv.FormatUint(pc.ID, 10)

	pc.Extra.FilterLabels[apistructs.LabelPipelineCronCompensated] = "true"

	return s.CreateV2(&apistructs.PipelineCreateRequestV2{
		PipelineYml:            pc.Extra.PipelineYml,
		ClusterName:            pc.Extra.ClusterName,
		PipelineYmlName:        pc.PipelineYmlName,
		PipelineSource:         pc.PipelineSource,
		Labels:                 pc.Extra.FilterLabels,
		NormalLabels:           pc.Extra.NormalLabels,
		Envs:                   pc.Extra.Envs,
		ConfigManageNamespaces: pc.Extra.ConfigManageNamespaces,
		AutoRunAtOnce:          false,
		AutoStartCron:          false,
		IdentityInfo: apistructs.IdentityInfo{
			UserID:         pc.Extra.NormalLabels[apistructs.LabelUserID],
			InternalClient: "system-cron-compensator",
		},
	})
}
