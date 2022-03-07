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
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/queuemanager/types"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func TestParsePipelineIDFromWatchedKey(t *testing.T) {
	key := "/devops/pipeline/reconciler/345"
	pipelineID, err := parsePipelineIDFromWatchedKey(key)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(345), pipelineID)
}

func TestUpdateStatusBeforeReconcile(t *testing.T) {
	r := &Reconciler{}
	p := spec.Pipeline{
		PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusRunning},
	}
	err := r.updateStatusBeforeReconcile(p)
	assert.Equal(t, nil, err)
}

func TestMakeContextForPipelineReconcile(t *testing.T) {
	pCtx := makeContextForPipelineReconcile(1)
	cancel, ok := pCtx.Value(ctxKeyPipelineExitChCancelFunc).(context.CancelFunc)
	assert.Equal(t, true, ok)
	cancel()
}

func TestMakePipelineWatchedKey(t *testing.T) {
	key := makePipelineWatchedKey(1)
	assert.Equal(t, "/devops/pipeline/reconciler/1", key)
}

func TestListen(t *testing.T) {
	r := &Reconciler{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(r), "Listen", func(r *Reconciler, ctx context.Context) {
		return
	})
	defer pm1.Unpatch()
	t.Run("listen", func(t *testing.T) {
		r.Listen(context.Background())
	})
}

func TestReconciler_doWatch(t *testing.T) {
	type fields struct {
		PutPipelineIntoQueue struct {
			err            error
			needRetryIfErr bool
			popCh          <-chan struct{}
		}
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test_processingPipelines",
			fields: fields{
				PutPipelineIntoQueue: struct {
					err            error
					needRetryIfErr bool
					popCh          <-chan struct{}
				}{err: fmt.Errorf("error"), needRetryIfErr: true, popCh: nil},
			},
			args: args{
				key: etcdReconcilerWatchPrefix + "1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				processingPipelines: sync.Map{},
			}

			ctx := context.Background()
			queueManager := mockQueueManager{
				putPipelineIntoQueue: PutPipelineIntoQueue{
					err:            tt.fields.PutPipelineIntoQueue.err,
					needRetryIfErr: tt.fields.PutPipelineIntoQueue.needRetryIfErr,
					popCh:          tt.fields.PutPipelineIntoQueue.popCh,
				},
			}
			r.QueueManager = queueManager

			patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(r), "Add", func(r *Reconciler, pipelineID uint64) {
				return
			})
			defer patch1.Unpatch()

			r.Reconcile(ctx, tt.args.key)
		})
	}
}

type PutPipelineIntoQueue struct {
	err            error
	needRetryIfErr bool
	popCh          <-chan struct{}
}

type mockQueueManager struct {
	putPipelineIntoQueue PutPipelineIntoQueue
}

func (m mockQueueManager) IdempotentAddQueue(pq *apistructs.PipelineQueue) types.Queue {
	panic("implement me")
}

func (m mockQueueManager) QueryQueueUsage(pq *apistructs.PipelineQueue) *pb.QueueUsage {
	panic("implement me")
}

func (m mockQueueManager) PutPipelineIntoQueue(pipelineID uint64) (popCh <-chan struct{}, needRetryIfErr bool, err error) {
	return m.putPipelineIntoQueue.popCh, m.putPipelineIntoQueue.needRetryIfErr, m.putPipelineIntoQueue.err
}

func (m mockQueueManager) PopOutPipelineFromQueue(pipelineID uint64) {
	panic("implement me")
}

func (m mockQueueManager) BatchUpdatePipelinePriorityInQueue(pq *apistructs.PipelineQueue, pipelineIDs []uint64) error {
	panic("implement me")
}

func (m mockQueueManager) Stop() {
	panic("implement me")
}

func (m mockQueueManager) SendQueueToEtcd(queueID uint64) {
	panic("implement me")
}

func (m mockQueueManager) ListenInputQueueFromEtcd(ctx context.Context) {
	panic("implement me")
}

func (m mockQueueManager) SendUpdatePriorityPipelineIDsToEtcd(queueID uint64, pipelineIDS []uint64) {
	panic("implement me")
}

func (m mockQueueManager) ListenUpdatePriorityPipelineIDsFromEtcd(ctx context.Context) {
	panic("implement me")
}

func (m mockQueueManager) SendPopOutPipelineIDToEtcd(pipelineID uint64) {
	panic("implement me")
}

func (m mockQueueManager) ListenPopOutPipelineIDFromEtcd(ctx context.Context) {
	panic("implement me")
}

func (m mockQueueManager) Export() json.RawMessage {
	panic("implement me")
}

func (m mockQueueManager) Import(message json.RawMessage) error {
	panic("implement me")
}
