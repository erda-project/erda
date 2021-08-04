// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
