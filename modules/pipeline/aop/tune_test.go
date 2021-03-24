package aop

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/services/reportsvc"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func init() {
	bdl := bundle.New(bundle.WithAllAvailableClients())
	dbClient, _ := dbclient.New()
	report := reportsvc.New(reportsvc.WithDBClient(dbClient))

	Initialize(bdl, dbClient, report)
}

func TestHandlePipeline(t *testing.T) {
	ctx := NewContextForTask(spec.PipelineTask{ID: 1}, spec.Pipeline{}, aoptypes.TuneTriggerTaskBeforeWait)
	err := Handle(ctx)
	assert.NoError(t, err)

	// pipeline end
	ctx = NewContextForPipeline(spec.Pipeline{
		PipelineBase: spec.PipelineBase{ID: 1},
	}, aoptypes.TuneTriggerPipelineAfterExec)
	err = Handle(ctx)
	assert.NoError(t, err)
}

func TestHandleTask(t *testing.T) {
	// task end
	ctx := NewContextForTask(
		spec.PipelineTask{ID: 1, Status: apistructs.PipelineStatusSuccess},
		spec.Pipeline{
			PipelineBase: spec.PipelineBase{ID: 1},
		}, aoptypes.TuneTriggerTaskAfterExec,
	)
	err := Handle(ctx)
	assert.NoError(t, err)
}
