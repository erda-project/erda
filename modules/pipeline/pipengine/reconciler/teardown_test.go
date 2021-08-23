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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/modules/pipeline/spec"
)

func TestClosePipelineExitChannel(t *testing.T) {
	p := &spec.Pipeline{PipelineBase: spec.PipelineBase{ID: 1}}

	// empty context
	emptyCtx := context.TODO()
	closePipelineExitChannel(emptyCtx, p.ID)

	// context with exitCh
	exitCh := make(chan struct{})
	go func() {
		select {
		case <-exitCh:
			// do nothing
		}
	}()
	rightCtx := context.WithValue(context.TODO(), ctxKeyPipelineExitCh, exitCh)
	closePipelineExitChannel(rightCtx, p.ID)
}

func TestEmptyContextValue(t *testing.T) {
	ctx := context.TODO()

	_, ok := ctx.Value("key").(bool)
	assert.False(t, ok)
}
