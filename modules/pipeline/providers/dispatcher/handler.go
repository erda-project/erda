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
	"fmt"
	"strconv"
	"time"

	"github.com/erda-project/erda-infra/pkg/strutil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker/worker"
)

func (p *provider) continueDispatcher(ctx context.Context) {
	for pipelineID := range p.pipelineIDsChan {
		go func(pipelineID uint64) {
			w, err := p.pickOneWorker(ctx, pipelineID)
			if err != nil {
				p.Log.Errorf("failed to pick worker(need retry), pipelineID: %d, err: %v", pipelineID, err)
				p.Dispatch(ctx, pipelineID)
				return
			}
			// TODO data
			taskID := strutil.String(pipelineID)
			if err := p.Lw.Dispatch(ctx, w, strutil.String(pipelineID), nil); err != nil {
				p.Log.Errorf("failed to dispatch logic task to worker(need retry), taskID: %s, workerID: %s, err: %v", w.ID().String(), taskID, err)
				p.Dispatch(ctx, pipelineID)
				return
			}
		}(pipelineID)
	}
}

func (p *provider) loadRunningPipelines(ctx context.Context) {
	pipelineIDs, err := p.dbClient.ListPipelineIDsByStatuses(apistructs.ReconcilerRunningStatuses()...)
	if err != nil {
		p.Log.Errorf("failed to load running pipelines(need retry), err: %v", err)
		time.Sleep(p.Cfg.IntervalOfLoadRunningPipelines)
		p.loadRunningPipelines(ctx)
		return
	}
	for _, pipelineID := range pipelineIDs {
		p.Dispatch(ctx, pipelineID)
	}
	p.Log.Info("load running pipelines success")
}

func (p *provider) onWorkerAdd(ctx context.Context, ev leaderworker.Event) {
	p.consistent.Add(worker.New(worker.WithID(ev.WorkerID)))
}

func (p *provider) onWorkerDelete(ctx context.Context, ev leaderworker.Event) {
	p.consistent.Remove(ev.WorkerID.String())
}

func (p *provider) pickOneWorker(ctx context.Context, pipelineID uint64) (worker.Worker, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	locateMember := p.consistent.LocateKey([]byte(strutil.String(pipelineID)))
	w, err := p.Lw.GetWorker(ctx, worker.ID(locateMember.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to get picked worker, workerID: %s, err: %v", w.ID().String(), err)
	}
	return w, nil
}

func (p *provider) getPipelineIDFromEtcdKey(key string) (uint64, error) {
	if !strutil.HasPrefixes(key, p.Cfg.EtcdKeyPrefix) {
		return 0, fmt.Errorf("invalid key without special prefix, key: %s", key)
	}
	idstr := strutil.TrimPrefixes(key, p.Cfg.EtcdKeyPrefix)
	pipelienID, err := strconv.ParseUint(idstr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid key after trim prefix, idstr: %s, err: %v", idstr, err)
	}
	return pipelienID, nil
}
