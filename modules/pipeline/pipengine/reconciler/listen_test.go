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
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
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
