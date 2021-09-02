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

package pipengine

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler"
)

type Engine struct {
	mgr        *actionexecutor.Manager
	dbClient   *dbclient.Client
	reconciler *reconciler.Reconciler
}

const logPrefixContinueLoading = "continue load running pipelines"

var once sync.Once
var e Engine

func New(
	dbClient *dbclient.Client,
) *Engine {
	e = Engine{
		dbClient: dbClient,
	}
	return &e
}

func (engine *Engine) OnceDo(
	r *reconciler.Reconciler,
) error {
	var onceErr error

	once.Do(func() {
		//执行engine的一些初始化过程
		_, cfgChan, err := engine.dbClient.ListPipelineConfigsOfActionExecutor()
		if err != nil {
			onceErr = err
		}

		mgr := actionexecutor.GetManager()
		if err := mgr.Initialize(cfgChan); err != nil {
			onceErr = err
			return
		}

		engine.reconciler = r
		engine.mgr = mgr
	})
	return onceErr
}

func (engine *Engine) StartReconciler(ctx context.Context) {
	go engine.reconciler.Listen()
	go engine.reconciler.ContinueBackupThrottler()

	// 开始 Listen 后再开始加载已经在处理中的流水线，否则组件还未准备好，包括 eventManger(阻塞)
	go func() {
		time.Sleep(time.Second * 2)
		if err := engine.loadRunningPipelines(); err != nil {
			logrus.Errorf("engine: failed to load running pipelines, err: %v", err)
		}
	}()
}

func (engine *Engine) StartGC(ctx context.Context) {
	go engine.reconciler.ListenGC()
	go engine.reconciler.PipelineDatabaseGC()
	go engine.reconciler.CompensateGCNamespaces()
}

func (engine *Engine) Send(pipelineID uint64) {
	engine.reconciler.Add(pipelineID)
}

func (engine *Engine) WaitDBGC(pipelineID uint64, ttl uint64, needArchive bool) {
	engine.reconciler.WaitDBGC(pipelineID, ttl, needArchive)
}

// loadRunningPipelines load running pipeline from db.
func (engine *Engine) loadRunningPipelines() error {
	logrus.Infof("%s: begin load running pipelines", logPrefixContinueLoading)
	pipelineIDs, err := engine.dbClient.ListPipelineIDsByStatuses(apistructs.ReconcilerRunningStatuses()...)
	if err != nil {
		return err
	}

	// send pipeline id by interval time instead of at once
	total := len(pipelineIDs)
	intervalSec := time.Duration(conf.InitializeSendRunningIntervalSec())
	intervalNum := conf.InitializeSendRunningIntervalNum()
	maxTimes := total / int(intervalNum)
	for i := 0; i <= maxTimes; i++ {
		time.Sleep(intervalSec)
		end := (i + 1) * int(intervalNum)
		if end > total {
			end = total
		}
		for _, id := range pipelineIDs[i*int(intervalNum) : end] {
			go func(pipelineID uint64) {
				engine.Send(pipelineID)
				logrus.Debugf("%s: load running pipeline success, pipelineID: %d", logPrefixContinueLoading, pipelineID)
			}(id)
		}
	}
	logrus.Infof("%s: pipengine end load running pipelines", logPrefixContinueLoading)
	return nil
}
