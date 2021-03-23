package reconciler

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func TestGetSchedulableTasks(t *testing.T) {
	r := Reconciler{
		processingTasks: sync.Map{},
	}
	tests := []struct {
		name  string
		p     *spec.Pipeline
		tasks []*spec.PipelineTask

		expectSchedulableLen int
	}{
		{
			name: "only t1",
			p:    &spec.Pipeline{PipelineBase: spec.PipelineBase{ID: 1}},
			tasks: []*spec.PipelineTask{
				{Name: "t1", Status: apistructs.PipelineStatusAnalyzed, Extra: spec.PipelineTaskExtra{RunAfter: []string{}}},
			},
			expectSchedulableLen: 1,
		},
		{
			name: "t1 done, t2 depends on t1",
			p:    &spec.Pipeline{PipelineBase: spec.PipelineBase{ID: 1}},
			tasks: []*spec.PipelineTask{
				{Name: "t1", Status: apistructs.PipelineStatusSuccess, Extra: spec.PipelineTaskExtra{RunAfter: []string{}}},
				{Name: "t2", Status: apistructs.PipelineStatusAnalyzed, Extra: spec.PipelineTaskExtra{RunAfter: []string{"t1"}}},
			},
			expectSchedulableLen: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schedulableTasks, err := r.getSchedulableTasks(test.p, test.tasks)
			assert.NoError(t, err)
			assert.Equal(t, test.expectSchedulableLen, len(schedulableTasks))
		})
	}
}
