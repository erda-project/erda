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

package queuemanager

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/coreos/etcd/clientv3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/events"
	"github.com/erda-project/erda/modules/pipeline/providers/reconciler/legacy/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/strutil"
)

func getPipelineIDFromIncomingListenKey(key string, prefix string) (uint64, error) {
	idstr := strutil.TrimPrefixes(key, prefix)
	pipelineID, err := strconv.ParseUint(idstr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse pipeline id from incoming listen key, key: %s, err: %v", key, err)
	}
	return pipelineID, nil
}
func makeIncomingPipelineListenKey(prefix string, pipelineID uint64) string {
	return filepath.Join(prefix, strutil.String(pipelineID))
}

// listenIncomingPipeline listen incoming pipeline id
func (q *provider) listenIncomingPipeline(ctx context.Context) {
	q.Lw.ListenPrefix(ctx, q.Cfg.IncomingPipelineCfg.ListenPrefixWithSlash,
		func(ctx context.Context, event *clientv3.Event) {
			pipelineID, err := getPipelineIDFromIncomingListenKey(string(event.Kv.Key), q.Cfg.IncomingPipelineCfg.ListenPrefixWithSlash)
			if err != nil {
				q.Log.Errorf("failed to handle incoming pipeline(no retry), err: %v", err)
				return
			}
			q.addIncomingPipelineIntoQueue(pipelineID)
		},
		nil,
	)

}

func (q *provider) addIncomingPipelineIntoQueue(pipelineID uint64) {
	// add into queue
	popCh, needRetryIfErr, err := q.QueueManager.PutPipelineIntoQueue(pipelineID)
	if err != nil {
		rlog.PErrorf(pipelineID, "failed to put pipeline into queue")
		if needRetryIfErr {
			q.addIncomingPipelineIntoQueue(pipelineID)
			return
		}
		// no need retry, treat as failed
		q.treatPipelineFailed(pipelineID)
		return
	}
	rlog.PInfof(pipelineID, "added into queue, waiting to pop from the queue")
	go func() {
		<-popCh
		rlog.PInfof(pipelineID, "pop from the queue, begin dispatch")
		q.Dispatcher.Dispatch(context.Background(), pipelineID)
	}()
}

func (q *provider) treatPipelineFailed(pipelineID uint64) {
	p := spec.Pipeline{
		PipelineBase: spec.PipelineBase{
			ID:     pipelineID,
			Status: apistructs.PipelineStatusFailed,
		},
	}
	if err := q.dbClient.UpdatePipelineBaseStatus(p.ID, p.Status); err != nil {
		go q.treatPipelineFailed(pipelineID)
		return
	}
	// TODO get user id from etcd value
	events.EmitPipelineInstanceEvent(&p, p.GetRunUserID())
}
