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

	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker/worker"
	"github.com/erda-project/erda/pkg/strutil"
)

func (p *provider) onWorkerAdd(ctx context.Context, ev leaderworker.Event) {
	p.Log.Infof("worker added, refresh consistent, workerID: %s", ev.WorkerID)
	p.consistent.Add(worker.New(worker.WithID(ev.WorkerID)))
	// no need to relocate tasks which already dispatched
}

func (p *provider) onWorkerDelete(ctx context.Context, ev leaderworker.Event) {
	p.Log.Infof("worker deleted, refresh consistent, workerID: %s", ev.WorkerID)
	p.consistent.Remove(ev.WorkerID.String())
	// dispatch tasks belong to deleted worker
	p.dispatchingIDs.Range(func(key, value interface{}) bool {
		if value == ev.WorkerID.String() {
			// set not dispatching now
			p.dispatchingIDs.Delete(key)
			// dispatching again
			p.Dispatch(ctx, key.(uint64))
		}
		return true
	})
}

func (p *provider) pickOneWorker(ctx context.Context, pipelineID uint64) (worker.ID, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	var members []string
	for _, member := range p.consistent.GetMembers() {
		members = append(members, member.String())
	}
	locateMember := p.consistent.LocateKey([]byte(strutil.String(pipelineID)))
	if locateMember == nil {
		return "", fmt.Errorf("failed to find proper worker, pipelineID: %d, consistent members: %v", pipelineID, members)
	}
	workerID := worker.ID(locateMember.String())
	return workerID, nil
}
