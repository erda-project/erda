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

package leaderworker

import (
	"context"
	"path/filepath"

	"github.com/coreos/etcd/clientv3"

	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker/worker"
	"github.com/erda-project/erda/pkg/strutil"
)

func (p *provider) notifyWorkerAdd(ctx context.Context, w worker.Worker, typ worker.Type) error {
	workerBytes, err := w.MarshalJSON()
	if err != nil {
		return err
	}

	// report heartbeat before add
	if err := p.workerOnceReportHeartbeat(ctx, w); err != nil {
		return err
	}

	var ops []clientv3.Op
	switch typ {
	case worker.Candidate:
		ops = append(ops,
			clientv3.OpPut(p.makeEtcdWorkerKey(w.GetID(), worker.Candidate), string(workerBytes)),
		)
	case worker.Official:
		ops = append(ops,
			clientv3.OpDelete(p.makeEtcdWorkerKey(w.GetID(), worker.Candidate)),
			clientv3.OpPut(p.makeEtcdWorkerKey(w.GetID(), worker.Official), string(workerBytes)),
		)
	}

	_, err = p.EtcdClient.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		p.Log.Errorf("failed to notify worker add, workerID: %s, err: %v", w.GetID(), err)
		return err
	}

	return nil
}

func (p *provider) deleteWorker(ctx context.Context, w worker.Worker) error {
	key := p.makeEtcdWorkerKey(w.GetID(), w.GetType())
	_, err := p.EtcdClient.Delete(ctx, key)
	return err
}

func (p *provider) makeEtcdWorkerKey(workerID worker.ID, typ worker.Type) string {
	keyPrefix := p.makeEtcdWorkerKeyPrefix(typ)
	key := filepath.Join(keyPrefix, workerID.String())
	return key
}

func (p *provider) makeEtcdWorkerKeyPrefix(typ worker.Type) string {
	return filepath.Clean(filepath.Join(p.Cfg.Worker.EtcdKeyPrefixWithSlash, "type", typ.String())) + "/"
}

func (p *provider) getWorkerIDFromEtcdWorkerKey(key string, typ worker.Type) worker.ID {
	prefix := p.makeEtcdWorkerKeyPrefix(typ)
	return worker.ID(strutil.TrimPrefixes(key, prefix))
}

func (p *provider) makeEtcdWorkerGeneralDispatchPrefix() string {
	return filepath.Join(p.Cfg.Worker.EtcdKeyPrefixWithSlash, "dispatch/worker") + "/"
}

func (p *provider) makeEtcdWorkerLogicTaskListenPrefix(workerID worker.ID) string {
	prefix := p.makeEtcdWorkerGeneralDispatchPrefix()
	return filepath.Join(prefix, workerID.String(), "task") + "/"
}

// $prefix/worker/dispatch/worker/$workerID/task/$logicTaskID(such as: pipelineID)
func (p *provider) makeEtcdWorkerTaskDispatchKey(workerID worker.ID, logicTaskID worker.LogicTaskID) string {
	prefix := p.makeEtcdWorkerLogicTaskListenPrefix(workerID)
	return filepath.Join(prefix, logicTaskID.String())
}

func (p *provider) makeEtcdWorkerHeartbeatKeyPrefix() string {
	return filepath.Clean(filepath.Join(p.Cfg.Worker.EtcdKeyPrefixWithSlash, "heartbeat")) + "/"
}

func (p *provider) makeEtcdWorkerHeartbeatKey(workerID worker.ID) string {
	prefix := p.makeEtcdWorkerHeartbeatKeyPrefix()
	return filepath.Join(prefix, workerID.String())
}

func (p *provider) getWorkerIDFromEtcdWorkerHeartbeatKey(key string) worker.ID {
	prefix := p.makeEtcdWorkerHeartbeatKeyPrefix()
	return worker.ID(strutil.TrimPrefixes(key, prefix))
}
