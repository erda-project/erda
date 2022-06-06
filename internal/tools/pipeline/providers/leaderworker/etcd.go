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

	worker2 "github.com/erda-project/erda/internal/tools/pipeline/providers/leaderworker/worker"
	"github.com/erda-project/erda/pkg/strutil"
)

func (p *provider) deleteWorker(ctx context.Context, w worker2.Worker) error {
	key := p.makeEtcdWorkerKey(w.GetID(), w.GetType())
	_, err := p.EtcdClient.Delete(ctx, key)
	return err
}

func (p *provider) makeEtcdWorkerKey(workerID worker2.ID, typ worker2.Type) string {
	keyPrefix := p.makeEtcdWorkerKeyPrefix(typ)
	key := filepath.Join(keyPrefix, workerID.String())
	return key
}

func (p *provider) makeEtcdWorkerKeyPrefix(typ worker2.Type) string {
	return filepath.Clean(filepath.Join(p.Cfg.Worker.EtcdKeyPrefixWithSlash, "type", typ.String())) + "/"
}

func (p *provider) getWorkerIDFromEtcdWorkerKey(key string, typ worker2.Type) worker2.ID {
	prefix := p.makeEtcdWorkerKeyPrefix(typ)
	return worker2.ID(strutil.TrimPrefixes(key, prefix))
}

func (p *provider) makeEtcdWorkerGeneralDispatchPrefix() string {
	return filepath.Join(p.Cfg.Worker.EtcdKeyPrefixWithSlash, "dispatch/worker") + "/"
}

func (p *provider) makeEtcdWorkerLogicTaskListenPrefix(workerID worker2.ID) string {
	prefix := p.makeEtcdWorkerGeneralDispatchPrefix()
	return filepath.Join(prefix, workerID.String(), "task") + "/"
}

// $prefix/worker/dispatch/worker/$workerID/task/$logicTaskID(such as: pipelineID)
func (p *provider) makeEtcdWorkerTaskDispatchKey(workerID worker2.ID, logicTaskID worker2.LogicTaskID) string {
	prefix := p.makeEtcdWorkerLogicTaskListenPrefix(workerID)
	return filepath.Join(prefix, logicTaskID.String())
}

func (p *provider) makeEtcdLeaderLogicTaskCancelKey(logicTaskID worker2.LogicTaskID) string {
	return filepath.Join(p.makeEtcdLeaderLogicTaskCancelListenPrefix(), logicTaskID.String())
}
func (p *provider) makeEtcdLeaderLogicTaskCancelListenPrefix() string {
	return filepath.Join(p.Cfg.Leader.EtcdKeyPrefixWithSlash, "dispatch/cancel-task") + "/"
}
func (p *provider) getLogicTaskIDFromLeaderCancelKey(key string) worker2.LogicTaskID {
	return worker2.LogicTaskID(strutil.TrimPrefixes(key, p.makeEtcdLeaderLogicTaskCancelListenPrefix()))
}

func (p *provider) makeEtcdWorkerLogicTaskCancelListenPrefix(workerID worker2.ID) string {
	return filepath.Join(p.makeEtcdWorkerGeneralDispatchPrefix(), workerID.String(), "cancel-task") + "/"
}
func (p *provider) makeEtcdWorkerLogicTaskCancelKey(workerID worker2.ID, logicTaskID worker2.LogicTaskID) string {
	return filepath.Join(p.makeEtcdWorkerLogicTaskCancelListenPrefix(workerID), logicTaskID.String())
}
func (p *provider) getLogicTaskIDFromWorkerCancelKey(workerID worker2.ID, key string) worker2.LogicTaskID {
	prefix := p.makeEtcdWorkerLogicTaskCancelListenPrefix(workerID)
	if !strutil.HasPrefixes(key, prefix) {
		return ""
	}
	return worker2.LogicTaskID(strutil.TrimPrefixes(key, prefix))
}

func (p *provider) makeEtcdWorkerHeartbeatKeyPrefix() string {
	return filepath.Clean(filepath.Join(p.Cfg.Worker.EtcdKeyPrefixWithSlash, "heartbeat")) + "/"
}

func (p *provider) makeEtcdWorkerHeartbeatKey(workerID worker2.ID) string {
	prefix := p.makeEtcdWorkerHeartbeatKeyPrefix()
	return filepath.Join(prefix, workerID.String())
}

func (p *provider) getWorkerIDFromEtcdWorkerHeartbeatKey(key string) worker2.ID {
	prefix := p.makeEtcdWorkerHeartbeatKeyPrefix()
	return worker2.ID(strutil.TrimPrefixes(key, prefix))
}

// see: makeEtcdWorkerTaskDispatchKey
func (p *provider) getWorkerIDFromIncomingKey(key string) worker2.ID {
	prefix := p.makeEtcdWorkerGeneralDispatchPrefix()
	if !strutil.HasPrefixes(key, prefix) {
		return ""
	}
	workerIDAndSuffix := strutil.TrimPrefixes(key, prefix)
	workerIDAndLogicTaskID := strutil.Split(workerIDAndSuffix, "/task/")
	if len(workerIDAndLogicTaskID) != 2 {
		return ""
	}
	return worker2.ID(workerIDAndLogicTaskID[0])
}
func (p *provider) getWorkerIDFromWorkerLogicTaskCancelKey(key string) worker2.ID {
	prefix := p.makeEtcdWorkerGeneralDispatchPrefix()
	if !strutil.HasPrefixes(key, prefix) {
		return ""
	}
	workerIDAndSuffix := strutil.TrimPrefixes(key, prefix)
	workerIDAndLogicTaskID := strutil.Split(workerIDAndSuffix, "/cancel-task/")
	if len(workerIDAndLogicTaskID) != 2 {
		return ""
	}
	return worker2.ID(workerIDAndLogicTaskID[0])
}

// see: makeEtcdWorkerTaskDispatchKey
func (p *provider) getWorkerIDFromWorkerGeneralDispatchKey(key string) worker2.ID {
	// '/task/'
	workerID := p.getWorkerIDFromIncomingKey(key)
	if workerID != "" {
		return workerID
	}
	// '/cancel-task/'
	workerID = p.getWorkerIDFromWorkerLogicTaskCancelKey(key)
	if workerID != "" {
		return workerID
	}
	return ""
}

func (p *provider) getWorkerLogicTaskIDFromIncomingKey(workerID worker2.ID, key string) worker2.LogicTaskID {
	prefix := p.makeEtcdWorkerLogicTaskListenPrefix(workerID)
	return worker2.LogicTaskID(strutil.TrimPrefixes(key, prefix))
}
