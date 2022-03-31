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
	"sort"
	"sync"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/commonutil/costtimeutil"
	"github.com/erda-project/erda/modules/pipeline/commonutil/statusutil"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/events"
	"github.com/erda-project/erda/modules/pipeline/metrics"
	"github.com/erda-project/erda/modules/pipeline/providers/cron/compensator"
	"github.com/erda-project/erda/modules/pipeline/providers/reconciler/rutil"
	"github.com/erda-project/erda/modules/pipeline/providers/reconciler/schedulabletask"
	"github.com/erda-project/erda/modules/pipeline/providers/resourcegc"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/strutil"
)

// PipelineReconciler is reconciler for pipeline.
type PipelineReconciler interface {
	// IsReconcileDone check if reconciler is done.
	IsReconcileDone(ctx context.Context, p *spec.Pipeline) bool

	// NeedReconcile check whether this pipeline need reconcile.
	NeedReconcile(ctx context.Context, p *spec.Pipeline) bool

	// PrepareBeforeReconcile do something before reconcile.
	PrepareBeforeReconcile(ctx context.Context, p *spec.Pipeline) error

	// GetTasksCanBeConcurrentlyScheduled get all tasks which can be concurrently scheduled.
	GetTasksCanBeConcurrentlyScheduled(ctx context.Context, p *spec.Pipeline) ([]*spec.PipelineTask, error)

	// ReconcileSchedulableTasks reconcile schedulable tasks belong to one pipeline.
	ReconcileSchedulableTasks(ctx context.Context, p *spec.Pipeline, tasks []*spec.PipelineTask) error

	// UpdateCurrentReconcileStatusIfNecessary calculate current reconcile status and update if necessary.
	UpdateCurrentReconcileStatusIfNecessary(ctx context.Context, p *spec.Pipeline) error

	// TeardownAfterReconcileDone teardown one pipeline after reconcile done.
	TeardownAfterReconcileDone(ctx context.Context, p *spec.Pipeline)
}

type defaultPipelineReconciler struct {
	log             logs.Logger
	st              schedulabletask.Interface
	resourceGC      resourcegc.Interface
	cronCompensator compensator.Interface
	r               *provider

	// internal fields
	dbClient                                     *dbclient.Client
	processingTasks                              sync.Map
	defaultRetryInterval                         time.Duration
	calculatedPipelineStatusByAllReconciledTasks apistructs.PipelineStatus
}

func (pr *defaultPipelineReconciler) IsReconcileDone(ctx context.Context, p *spec.Pipeline) bool {
	return !pr.NeedReconcile(ctx, p)
}

func (pr *defaultPipelineReconciler) NeedReconcile(ctx context.Context, p *spec.Pipeline) bool {
	return !p.Status.IsEndStatus()
}

func (pr *defaultPipelineReconciler) PrepareBeforeReconcile(ctx context.Context, p *spec.Pipeline) error {
	// update pipeline status if necessary
	// send event in a tx
	if p.Status.AfterPipelineQueue() {
		return nil
	}
	//_, err := pr.dbClient.Transaction(func(s *xorm.Session) (interface{}, error) {
	// update status
	if err := pr.dbClient.UpdatePipelineBaseStatus(p.ID, apistructs.PipelineStatusRunning); err != nil {
		return err
	}
	pr.log.Infof("pipelineID: %d, update pipeline status (%s -> %s)", p.ID, p.Status, apistructs.PipelineStatusRunning)
	p.Status = apistructs.PipelineStatusRunning
	// send event
	events.EmitPipelineInstanceEvent(p, p.GetUserID())
	return nil
	//})
	//return err
}

func (pr *defaultPipelineReconciler) GetTasksCanBeConcurrentlyScheduled(ctx context.Context, p *spec.Pipeline) ([]*spec.PipelineTask, error) {
	// get all tasks
	allTasks, err := pr.r.ymlTaskMergeDBTasks(p)
	if err != nil {
		return nil, err
	}

	schedulableTasks, err := pr.st.GetSchedulableTasks(ctx, p, allTasks)
	if err != nil {
		return nil, err
	}
	var filteredTasks []*spec.PipelineTask
	for _, task := range schedulableTasks {
		_, onProcessing := pr.processingTasks.LoadOrStore(task.Name, struct{}{})
		if !onProcessing {
			filteredTasks = append(filteredTasks, task)
		}
	}

	// print
	var filteredTaskNames []string
	for _, task := range filteredTasks {
		filteredTaskNames = append(filteredTaskNames, task.Name)
	}
	sort.Strings(filteredTaskNames)
	pr.log.Infof("pipelineID: %d, schedulable tasks: %s", p.ID, strutil.Join(filteredTaskNames, ", ", true))

	return filteredTasks, nil
}

func (pr *defaultPipelineReconciler) ReconcileSchedulableTasks(ctx context.Context, p *spec.Pipeline, tasks []*spec.PipelineTask) error {
	var wg sync.WaitGroup
	wg.Add(len(tasks))
	for _, task := range tasks {
		task := task
		go func() {
			defer wg.Done()
			tr := &defaultTaskReconciler{
				log:                  pr.r.Log.Sub("task"),
				policy:               pr.r.TaskPolicy,
				cache:                pr.r.Cache,
				clusterInfo:          pr.r.ClusterInfo,
				r:                    pr.r,
				pr:                   pr,
				dbClient:             pr.dbClient,
				bdl:                  pr.r.bdl,
				defaultRetryInterval: pr.r.Cfg.RetryInterval,
				pipelineSvcFuncs:     pr.r.pipelineSvcFuncs,
				actionAgentSvc:       pr.r.actionAgentSvc,
				extMarketSvc:         pr.r.extMarketSvc,
			}
			tr.ReconcileOneTaskUntilDone(ctx, p, task)
		}()
	}
	wg.Wait()

	return nil
}

func (pr *defaultPipelineReconciler) UpdateCurrentReconcileStatusIfNecessary(ctx context.Context, p *spec.Pipeline) error {
	// fetch the latest pipeline after reconciled batch tasks for condition like StopByUser bypass reconciler.
	// TODO can remove after stop at reconciler side
	*p = *pr.r.mustFetchPipelineDetail(ctx, p.ID)
	if p.Status.IsEndStatus() {
		return nil
	}

	// get all tasks
	allTasks, err := pr.r.ymlTaskMergeDBTasks(p)
	if err != nil {
		return err
	}

	// set calculated pipeline status by all reconciled tasks from db
	if err := pr.calculatePipelineStatusByAllReconciledTasks(ctx, p); err != nil {
		return err
	}

	// all tasks need to be end status, and then update pipeline status
	allTasksDone := statusutil.CalculatePipelineTaskAllDone(allTasks)
	if !allTasksDone {
		return nil
	}

	// calculate pipeline status
	calculatedPipelineStatus := statusutil.CalculatePipelineStatusV2(allTasks)
	// no change, exit
	if p.Status == calculatedPipelineStatus {
		return nil
	}
	// changed, update pipeline status
	//_, err = pr.dbClient.Transaction(func(s *xorm.Session) (interface{}, error) {
	// update status
	if err := pr.dbClient.UpdatePipelineBaseStatus(p.ID, calculatedPipelineStatus); err != nil {
		return err
	}
	p.Status = calculatedPipelineStatus
	// send event
	events.EmitPipelineInstanceEvent(p, p.GetUserID())
	return nil
	//})
	//if err != nil {
	//	return err
	//}

	//return nil
}

func (pr *defaultPipelineReconciler) TeardownAfterReconcileDone(ctx context.Context, p *spec.Pipeline) {
	pr.log.Infof("begin teardown pipeline, pipelineID: %d", p.ID)
	defer pr.log.Infof("end teardown pipeline, pipelineID: %d", p.ID)

	// update end time
	now := time.Now()
	rutil.ContinueWorking(ctx, pr.log, func(ctx context.Context) rutil.WaitDuration {
		// already updated
		if p.TimeEnd != nil {
			return rutil.ContinueWorkingAbort
		}
		p.TimeEnd = &now
		p.CostTimeSec = costtimeutil.CalculatePipelineCostTimeSec(p)
		if err := pr.dbClient.UpdatePipelineBase(p.ID, &p.PipelineBase); err != nil {
			pr.log.Errorf("failed to update pipeline when teardown(auto retry), pipelineID: %d, err: %v", p.ID, err)
			return rutil.ContinueWorkingWithDefaultInterval
		}
		return rutil.ContinueWorkingAbort
	}, rutil.WithContinueWorkingDefaultRetryInterval(pr.defaultRetryInterval))

	// metrics
	go metrics.PipelineCounterTotalAdd(*p, 1)
	go metrics.PipelineGaugeProcessingAdd(*p, -1)
	go metrics.PipelineEndEvent(*p)
	// aop
	rutil.ContinueWorking(ctx, pr.log, func(ctx context.Context) rutil.WaitDuration {
		if err := aop.Handle(aop.NewContextForPipeline(*p, aoptypes.TuneTriggerPipelineAfterExec)); err != nil {
			pr.log.Errorf("failed to do aop at pipeline-after-exec, pipelineID: %d, err: %v", p.ID, err)
		}
		// TODO continue retry maybe block teardown if there is a bad aop plugin
		return rutil.ContinueWorkingAbort
	}, rutil.WithContinueWorkingDefaultRetryInterval(pr.defaultRetryInterval))

	// cron compensator
	pr.cronCompensator.PipelineCronCompensate(ctx, p.ID)
	// resource gc
	pr.resourceGC.WaitGC(p.Extra.Namespace, p.ID, p.GetResourceGCTTL())

	// mark teardown
	rutil.ContinueWorking(ctx, pr.log, func(ctx context.Context) rutil.WaitDuration {
		if p.Extra.CompleteReconcilerTeardown {
			return rutil.ContinueWorkingAbort
		}
		p.Extra.CompleteReconcilerTeardown = true
		if err := pr.dbClient.UpdatePipelineExtraByPipelineID(p.ID, &p.PipelineExtra); err != nil {
			pr.log.Errorf("failed to update pipeline complete teardown mark(auto retry), pipelineID: %d, err: %v)", p.ID, err)
			return rutil.ContinueWorkingWithCustomInterval(pr.r.Cfg.RetryInterval)
		}
		return rutil.ContinueWorkingAbort
	})
}

func (pr *defaultPipelineReconciler) calculatePipelineStatusByAllReconciledTasks(ctx context.Context, p *spec.Pipeline) error {
	reconciledTasks, err := pr.dbClient.ListPipelineTasksByPipelineID(p.ID)
	if err != nil {
		return err
	}
	var tasks []*spec.PipelineTask
	for _, t := range reconciledTasks {
		tasks = append(tasks, &t)
	}
	pr.calculatedPipelineStatusByAllReconciledTasks = statusutil.CalculatePipelineStatusV2(tasks)
	return nil
}
