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
	closePipelineExitChannel(emptyCtx, p)

	// context with exitCh
	exitCh := make(chan struct{})
	go func() {
		select {
		case <-exitCh:
			// do nothing
		}
	}()
	rightCtx := context.WithValue(context.TODO(), ctxKeyPipelineExitCh, exitCh)
	closePipelineExitChannel(rightCtx, p)
}

func TestEmptyContextValue(t *testing.T) {
	ctx := context.TODO()

	_, ok := ctx.Value("key").(bool)
	assert.False(t, ok)
}
