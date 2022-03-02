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

package dispatcher

import (
	"context"
	"time"

	"github.com/erda-project/erda-infra/pkg/strutil"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker/worker"
)

func (p *provider) continueDispatcher(ctx context.Context) {
	for {
		c, err := p.makeConsistent(ctx)
		if err != nil {
			p.Log.Errorf("failed to init consistent(need retry), err: %v", err)
			time.Sleep(p.Cfg.DispatchRetryInterval)
			continue
		}
		p.lock.Lock()
		p.consistent = c
		p.lock.Unlock()
		break
	}

	for pipelineID := range p.pipelineIDsChan {
		_, dispatching := p.dispatchingIDs.LoadOrStore(pipelineID, "")
		if dispatching {
			continue
		}
		go func(pipelineID uint64) {
			// pick one worker
			workerID, err := p.pickOneWorker(ctx, pipelineID)
			if err != nil {
				p.dispatchingIDs.Delete(pipelineID)
				p.Log.Errorf("failed to pick worker(need retry after %s), pipelineID: %d, err: %v", p.Cfg.DispatchRetryInterval, pipelineID, err)
				time.Sleep(p.Cfg.DispatchRetryInterval)
				p.Dispatch(ctx, pipelineID)
				return
			}
			// assign lock task to the picked worker
			logicTaskID := worker.TaskLogicID(strutil.String(pipelineID))
			logicTaskData := []byte(nil)
			if err := p.Lw.AssignLogicTaskToWorker(ctx, workerID, worker.NewTasker(logicTaskID, logicTaskData)); err != nil {
				p.dispatchingIDs.Delete(pipelineID)
				p.Log.Errorf("failed to dispatch logic task to worker(need retry after %s), taskID: %s, workerID: %s, err: %v", p.Cfg.DispatchRetryInterval, logicTaskID, workerID, err)
				time.Sleep(p.Cfg.DispatchRetryInterval)
				p.Dispatch(ctx, pipelineID)
				return
			}

			p.dispatchingIDs.Store(pipelineID, workerID.String())
			p.Log.Infof("assign logic task to worker success, pipelineID: %d, workerID: %s", pipelineID, workerID)
		}(pipelineID)
	}
}
