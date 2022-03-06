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

	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker/worker"
)

type Event struct {
	Type         mvccpb.Event_EventType
	WorkerID     worker.ID
	LogicTaskIDs []worker.TaskLogicID
}

type workerWithCancel struct {
	Worker     worker.Worker
	Ctx        context.Context
	CancelFunc context.CancelFunc
}

type (
	WorkerAddHandler    func(ctx context.Context, ev Event)
	WorkerDeleteHandler func(ctx context.Context, ev Event)
)
