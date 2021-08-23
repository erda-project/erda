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
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/commonutil/costtimeutil"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/loop"
)

func (r *Reconciler) teardownCurrentReconcile(ctx context.Context, pipelineID uint64) {
	closePipelineExitChannel(ctx, pipelineID)
	r.deleteEtcdWatchKey(context.Background(), pipelineID)
	r.teardownPipelines.Delete(pipelineID)
	r.QueueManager.PopOutPipelineFromQueue(pipelineID)
}

func (r *Reconciler) teardownPipeline(ctx context.Context, p *spec.PipelineWithTasks) {
	if !p.Pipeline.Status.IsEndStatus() {
		return
	}

	if _, ing := r.teardownPipelines.LoadOrStore(p.Pipeline.ID, true); ing {
		return
	}
	logrus.Infof("reconciler: begin teardown pipeline, pipelineID: %d", p.Pipeline.ID)
	defer r.doCronCompensate(ctx, p.Pipeline.ID)
	defer r.waitGC(p.Pipeline.Extra.Namespace, p.Pipeline.ID, p.Pipeline.GetResourceGCTTL())
	defer func() {
		// // metrics
		// go metrics.PipelineCounterTotalAdd(*p.Pipeline, 1)
		// go metrics.PipelineGaugeProcessingAdd(*p.Pipeline, -1)
		// go metrics.PipelineEndEvent(*p.Pipeline)
		// aop
		_ = aop.Handle(aop.NewContextForPipeline(*p.Pipeline, aoptypes.TuneTriggerPipelineAfterExec))
	}()
	defer logrus.Infof("reconciler: pipelineID: %d, pipeline is completed", p.Pipeline.ID)
	for _, task := range p.Tasks {
		if task.Status == apistructs.PipelineStatusAnalyzed || task.Status == apistructs.PipelineStatusBorn {
			task.Status = apistructs.PipelineStatusNoNeedBySystem
			if err := r.dbClient.UpdatePipelineTaskStatus(task.ID, task.Status); err != nil {
				logrus.Errorf("[alert] reconciler: pipelineID: %d, task: %s, failed to teardown pipeline (%v)",
					p.Pipeline.ID, task.Name, err)
			}
			continue
		}
	}
	// 推进完毕设置 snippet task 状态
	if err := r.fulfillParentSnippetTask(p.Pipeline); err != nil {
		logrus.Errorf("[alert] reconciler: pipelineID: %d, failed to teardown pipeline (failed to fulfillSnippetTask, err: %v)", p.Pipeline.ID, err)
	}
	// 更新结束时间
	now := time.Now()
	p.Pipeline.TimeEnd = &now
	p.Pipeline.CostTimeSec = costtimeutil.CalculatePipelineCostTimeSec(p.Pipeline)
	if err := r.dbClient.UpdatePipelineBase(p.Pipeline.ID, &p.Pipeline.PipelineBase); err != nil {
		logrus.Errorf("[alert] reconciler: pipelineID: %d, failed to teardown pipeline (failed to update pipeline: %v)",
			p.Pipeline.ID, err)
	}
	// 标记已完成 teardown
	p.Pipeline.Extra.CompleteReconcilerTeardown = true
	if err := r.dbClient.UpdatePipelineExtraByPipelineID(p.Pipeline.ID, &p.Pipeline.PipelineExtra); err != nil {
		logrus.Errorf("[alert] reconciler: pipelineID: %d, failed to teardown pipeline (failed to update pipeline complete teardown mark: %v)",
			p.Pipeline.ID, err)
	}
	logrus.Infof("reconciler: end teardown pipeline, pipelineID: %d", p.Pipeline.ID)
}

// closePipelineExitChannel send signal to pipeline exit channel to stop other related things.
func closePipelineExitChannel(ctx context.Context, pipelineID uint64) {
	rlog.PDebugf(pipelineID, "pipeline exit, begin send signal to exit channel to stop other related things")
	defer rlog.PDebugf(pipelineID, "pipeline exit, end send signal to exit channel to stop other related things")
	defer func() {
		if err := recover(); err != nil {
			rlog.PErrorf(pipelineID, "pipeline occurred a panic when closePipelineExitChannel, err: %v", err)
		}
	}()
	exitCh, ok := ctx.Value(ctxKeyPipelineExitCh).(chan struct{})
	if !ok {
		return
	}
	exitCh <- struct{}{}
	close(exitCh)
}

// deleteEtcdWatchKey delete pipeline corresponding etcd watched key.
func (r *Reconciler) deleteEtcdWatchKey(ctx context.Context, pipelineID uint64) {
	rlog.PDebugf(pipelineID, "start delete etcd watch key")
	defer rlog.PInfof(pipelineID, "end delete etcd watch key")

	etcdKey := makePipelineWatchedKey(pipelineID)
	_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Second*60)).Do(func() (abort bool, err error) {
		err = r.js.Remove(ctx, etcdKey, nil)
		if err != nil {
			return false, rlog.PErrorAndReturn(pipelineID, fmt.Errorf("failed to delete etcd watch key, err: %v", err))
		}
		rlog.PDebugf(pipelineID, "delete etcd watch key success")
		return true, nil
	})
}
