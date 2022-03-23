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

package compensator

import (
	"context"
	"fmt"
	"reflect"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/etcd"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/limit_sync_group"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	waitTimeIfQueryDBError   = time.Minute
	indexFirst               = 0
	EtcdNeedCompensatePrefix = "/devops/pipeline/compensate/"
)

type config struct {
}

// +provider
type provider struct {
	LeaderWorker leaderworker.Interface `autowired:"leader-worker"`
	ETCD         etcd.Interface         // autowired
	EtcdClient   *v3.Client
	MySQL        mysqlxorm.Interface `autowired:"mysql-xorm"`

	client       *dbclient.Client
	pipelineFunc PipelineFunc
}

func (p *provider) WithPipelineFunc(pipelineFunc PipelineFunc) {
	p.client = &dbclient.Client{Engine: p.MySQL.DB()}
	p.pipelineFunc = pipelineFunc
}

func (p *provider) Init(ctx servicehub.Context) error {
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	p.LeaderWorker.OnLeader(func(ctx context.Context) {
		// The client value may not have been set during initialization
		p.ContinueCompensate(context.Background())
	})
	return nil
}

var compensateLog = logrus.WithField("type", "cron compensator")

func getCronCompensateInterval(interval int64) time.Duration {
	return time.Duration(interval) * time.Minute
}

func getCronInterruptCompensateInterval(interval int64) time.Duration {
	return time.Duration(interval) * time.Minute * 2
}

func (p *provider) ContinueCompensate(ctx context.Context) {
	compensateLog.Info("cron compensator: start")
	// Execute Policy Compensation Scheduled Tasks
	ticker := time.NewTicker(getCronCompensateInterval(conf.CronCompensateTimeMinute()))
	// Interrupt Compensation Scheduled Task
	interruptTicker := time.NewTicker(getCronInterruptCompensateInterval(conf.CronCompensateTimeMinute()))

	// Execute an interruption compensation first when the project starts
	p.traverseDoCompensate(p.doInterruptCompensate, true)

	for {
		select {
		case <-ctx.Done():
			// stop
			compensateLog.Info("stop cron compensate, received cancel signal from channel")
			ticker.Stop()
			interruptTicker.Stop()
			return
		case <-interruptTicker.C:
			// Why synchronization is used for interrupt compensation here,
			// because data will be added here,
			// and it is not yet confirmed whether there is idempotent, so synchronize first
			p.traverseDoCompensate(p.doInterruptCompensate, true)
		case <-ticker.C:
			// Because the compensation (strategy) is not executed,
			// the pipeline used to execute the pipeline is idempotent internally,
			// that is, only one pipeline is executing at the same time
			p.traverseDoCompensate(p.doStrategyCompensate, false)
		}
	}
}

func (p *provider) doInterruptCompensate(pc spec.PipelineCron) {
	err := p.cronInterruptCompensate(pc)
	if err != nil {
		compensateLog.WithField("cronID", pc.ID).Errorf("failed to do interrupt-compensate, cronID: %d, err: %v", pc.ID, err)
	}
}

func (p *provider) doStrategyCompensate(pc spec.PipelineCron) {
	err := p.cronNotExecuteCompensate(pc)
	if err != nil {
		compensateLog.WithField("cronID", pc.ID).Errorf("failed to do notexecute-compensate, cronID: %d, err: %v", pc.ID, err)
	}
}

func (p *provider) traverseDoCompensate(doCompensate func(cron spec.PipelineCron), sync bool) {

	if doCompensate == nil {
		return
	}

	// get all enabled crons
	enabledCrons, err := p.client.ListPipelineCrons(&[]bool{true}[0])
	if err != nil {
		compensateLog.Errorf("failed to list enabled pipeline crons from db, try again later, err: %v", err)
		time.Sleep(waitTimeIfQueryDBError)
		return
	}

	group := limit_sync_group.NewSemaphore(int(conf.CronCompensateConcurrentNumber()))
	for _, pc := range enabledCrons {
		if p.isCronShouldBeIgnored(pc) {
			triggerTime := time.Now()
			logrus.Warnf("crond compensator: pipelineCronID: %d, triggered compensate but ignored, triggerTime: %s, cronStartFrom: %s",
				pc.ID, triggerTime, *pc.Extra.CronStartFrom)
			continue
		}
		if sync {
			doCompensate(pc)
		} else {
			group.Add(1)
			go func(pc spec.PipelineCron) {
				defer group.Done()
				doCompensate(pc)
			}(pc)
		}
	}
	group.Wait()
}

// cronInterruptCompensate Timing interrupt compensation
func (p *provider) cronInterruptCompensate(pc spec.PipelineCron) error {

	// Calculate interrupt compensation start time
	beforeCompensateFromTime := getCompensateFromTime(pc)

	// current time minus a certain time，prevent conflicts with cron create pipeline
	now := time.Unix(time.Now().Unix(), 0)
	var thisCompensateFromTime = now.Add(time.Second * -time.Duration(conf.CronFailureCreateIntervalCompensateTimeSecond()))

	// Use cron expr to calculate all required trigger times from the starting compensation point
	needTriggerTimes, err := pipelineyml.ListNextCronTime(pc.CronExpr,
		pipelineyml.WithCronStartEndTime(&beforeCompensateFromTime, &thisCompensateFromTime),
		pipelineyml.WithListNextScheduleCount(100),
	)
	if err != nil {
		return errors.Errorf("[alert] failed to list next crontimes, cronID: %d, err: %v", pc.ID, err)
	}

	if len(needTriggerTimes) > 0 {
		// Pipeline search record created based on createlyname + source
		existPipelines, _, _, _, err := p.client.PageListPipelines(apistructs.PipelinePageListRequest{
			Sources:          []apistructs.PipelineSource{pc.PipelineSource},
			YmlNames:         []string{pc.PipelineYmlName},
			TriggerModes:     []apistructs.PipelineTriggerMode{apistructs.PipelineTriggerModeCron},
			StartTimeCreated: beforeCompensateFromTime.Add(time.Second * -time.Duration(conf.CronFailureCreateIntervalCompensateTimeSecond())),
			EndTimeCreated:   thisCompensateFromTime.Add(time.Second * time.Duration(conf.CronFailureCreateIntervalCompensateTimeSecond())),
			PageNum:          1,
			PageSize:         100,
			LargePageSize:    true,
		})
		if err != nil {
			return errors.Errorf("[alert] failed to list existPipelines, cronID: %d, err: %v", pc.ID, err)
		}

		// Convert to map for query
		existPipelinesMap := make(map[time.Time]spec.Pipeline, len(existPipelines))
		for _, p := range existPipelines {
			existPipelinesMap[getTriggeredTime(p)] = p
		}

		// Traverse needTriggerTimes. If it is not created, you need to interrupt the creation of compensation
		for _, ntt := range needTriggerTimes {
			pipeline, ok := existPipelinesMap[ntt]
			if ok {
				compensateLog.Infof("no need do interrupt-compensate, cronID: %d, triggerTime: %v, exist pipelineID: %d", pc.ID, ntt, pipeline.ID)
				continue
			}
			compensateLog.Infof("need do interrupt-compensate, cronID: %d, triggerTime: %v", pc.ID, ntt)
			// create
			created, err := p.createCronCompensatePipeline(pc, ntt)
			if err != nil {
				compensateLog.Errorf("failed to do interrupt-compensate, cronID: %d, triggerTime: %v, err: %v", pc.ID, ntt, err)
				continue
			}
			compensateLog.Infof("success to do interrupt-compensate, cronID: %d, triggerTime: %v, createdPipelineID: %d", pc.ID, ntt, created.ID)
		}
	}

	// After the interrupt compensation is completed, the thisCompensateFromTime field of cron needs to be updated
	pc.Extra.LastCompensateAt = &thisCompensateFromTime
	// If the compensator is empty, it indicates that it is an old cron, and the default configuration will be used automatically
	if pc.Extra.Compensator == nil {
		pc.Extra.Compensator = &apistructs.CronCompensator{
			Enable:               pipelineyml.DefaultCronCompensator.Enable,
			LatestFirst:          pipelineyml.DefaultCronCompensator.LatestFirst,
			StopIfLatterExecuted: pipelineyml.DefaultCronCompensator.StopIfLatterExecuted,
		}
	}
	if err := p.client.UpdatePipelineCron(pc.ID, &pc); err != nil {
		return errors.Errorf("failed to update pipelineCron for lastCompensateAt field, err: %v", err)
	}

	return nil
}

func (p *provider) CronNotExecuteCompensateById(id uint64) error {
	cron, err := p.client.GetPipelineCron(id)
	if err != nil {
		return err
	}

	return p.cronNotExecuteCompensate(cron)
}

// cronNotExecuteCompensate timing compensation not performed
// Only within one day
func (p *provider) cronNotExecuteCompensate(pc spec.PipelineCron) error {

	// Notexecute compensate is not enabled, exit
	if pc.Enable == nil || *pc.Enable == false || pc.Extra.Compensator == nil || pc.Extra.Compensator.Enable == false {
		return nil
	}

	now := time.Unix(time.Now().Unix(), 0)
	oneDayBeforeNow := now.AddDate(0, 0, -1)

	// Get to execute list
	// Search the created pipeline records according to source + ymlname + ID
	// Here, why take 10 by ID? Because when the expression granularity is very small, a lot of data will be lost. However, the ID order of interrupt compensation is different from that of execution
	// Therefore, neutralization first takes 10 items in positive or reverse order with ID, and then doCronCompensate selects the most suitable time for execution according to the specific execution time of the 10 items
	// When the time granularity is very large, in essence, one is OK
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

	existPipelines, _, _, _, err := p.client.PageListPipelines(request)
	if err != nil {
		return errors.Errorf("failed to list notexecute pipelines, cronID: %d, err: %v", pc.ID, err)
	}

	return p.doCronCompensate(*pc.Extra.Compensator, existPipelines, pc)
}

func (p *provider) doCronCompensate(compensator apistructs.CronCompensator, notRunPipelines []spec.Pipeline, pipelineCron spec.PipelineCron) error {
	var order string

	if len(notRunPipelines) <= 0 {
		return nil
	}

	// Select the most suitable time point from the non execution in good order according to the strategy
	if compensator.LatestFirst {
		order = "DESC"
	} else {
		order = "ASC"
	}
	// doCronCompensate selects the most suitable time for execution according to the specific execution time of Article 10
	firstOrLastPipeline := orderByCronTriggerTime(notRunPipelines, order)[indexFirst]

	// According to the policy decision, if it is the last pipeline, when it is the StopIfLatterExecuted policy,
	// it should be compared with the pipeline in the latest success status. Only the ID greater than the successful ID can be executed
	if compensator.LatestFirst && compensator.StopIfLatterExecuted {
		// Get the pipeline successfully executed
		runSuccessPipeline, _, _, _, err := p.client.PageListPipelines(apistructs.PipelinePageListRequest{
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

		// If the latest successful ID is greater than the compensated ID, no compensation will be made
		lastSuccessPipeline := runSuccessPipeline[indexFirst]
		if lastSuccessPipeline.ID > firstOrLastPipeline.ID {
			return nil
		}
	}

	_, err := p.pipelineFunc.RunPipeline(&apistructs.PipelineRunRequest{
		PipelineID:   firstOrLastPipeline.ID,
		IdentityInfo: apistructs.IdentityInfo{InternalClient: firstOrLastPipeline.Extra.InternalClient},
	})

	// Print one line of record after successful execution
	if err == nil {
		compensateLog.Infof("[doCronCompensate] Compensate success, pipelineId %d", firstOrLastPipeline.ID)
		return nil
	}

	// If there is a conflict or internal error in compensation, the listener should be notified and the compensation scheduling should be directly executed next time,
	// If the sum of the corresponding cron expression interval and execution time is greater than the scheduling time without compensation,
	// the whole scheduling will be executed as the configured policy. If it is less than, there will be competitive execution
	compensateLog.Infof("[doCronCompensate] run Compensate err, put cronId into etcd wait callback: cronId %d", pipelineCron.ID)
	// Create etcd lease
	lease := v3.NewLease(p.EtcdClient)
	if grant, err := lease.Grant(context.Background(), conf.CronCompensateTimeMinute()*60); err == nil {
		// Set cronid to key and wait
		if _, err := p.EtcdClient.Put(context.Background(),
			fmt.Sprint(EtcdNeedCompensatePrefix, pipelineCron.ID),
			"", v3.WithLease(grant.ID)); err != nil {
			// Failed to write etcd. This compensation failed. Wait for the next compensation
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

// orderByCronTriggerTime order
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

// getCompensateFromTime calculates the start time point of interrupt compensation
// If extra LastCompensateAt is empty, which means it has not been compensated. The cron update time is used as the compensation start time
// The maximum compensation is one day
//
// There is a special scenario:
// 1.  Manually change the enable field to 0 and restart the pipeline to stop cron
// 2.  Do some operations that need to temporarily stop cron, such as database migration, cluster adjustment, etc
// 3.  Manually modify enable = 1, restart the pipeline to make cron effective, and interrupt compensation is required
// During the process, the LastCompensateAt field is not updated, and cron is temporarily stopped. In this case, the cron update time is also used as the compensation start time
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

// getTriggeredTime gets the creation time. The pipeline created regularly uses CronTriggerTime
func getTriggeredTime(p spec.Pipeline) time.Time {
	if p.Extra.CronTriggerTime != nil {
		return time.Unix((*p.Extra.CronTriggerTime).Unix(), 0)
	}
	return time.Unix(p.TimeCreated.Unix(), 0)
}

func (p *provider) createCronCompensatePipeline(pc spec.PipelineCron, triggerTime time.Time) (*spec.Pipeline, error) {
	// generate new label map avoid concurrent map problem
	pc.Extra.NormalLabels = pc.GenCompensateCreatePipelineReqNormalLabels(triggerTime)
	pc.Extra.FilterLabels = pc.GenCompensateCreatePipelineReqFilterLabels()

	return p.pipelineFunc.CreatePipeline(&apistructs.PipelineCreateRequestV2{
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

// isCronShouldIgnore if trigger time before cron start from time, should ignore cron at this trigger time
func (p *provider) isCronShouldBeIgnored(pc spec.PipelineCron) bool {
	if pc.Extra.CronStartFrom == nil {
		return false
	}
	triggerTime := time.Now()
	return triggerTime.Before(*pc.Extra.CronStartFrom)
}

func init() {
	servicehub.Register("cron-compensator", &servicehub.Spec{
		Services:   []string{"cron-compensator"},
		Types:      []reflect.Type{reflect.TypeOf((*Interface)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
