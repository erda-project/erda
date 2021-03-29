package taskrun

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/pkg/strutil"
)

// Teardown tear down task.
func (tr *TaskRun) Teardown() {
	logrus.Infof("reconciler: pipelineID: %d, task %q begin tear down", tr.P.ID, tr.Task.Name)
	defer logrus.Infof("reconciler: pipelineID: %d, task %q end tear down", tr.P.ID, tr.Task.Name)
	defer tr.TeardownConcurrencyCount()
	defer tr.TeardownPriorityQueue()
	defer aop.Handle(aop.NewContextForTask(*tr.Task, *tr.P, aoptypes.TuneTriggerTaskAfterExec))

	// invalidate openapi oauth2 token
	tokens := strutil.DedupSlice([]string{
		tr.Task.Extra.PublicEnvs[apistructs.EnvOpenapiTokenForActionBootstrap],
		tr.Task.Extra.PrivateEnvs[apistructs.EnvOpenapiToken],
	}, true)
	for _, token := range tokens {
		_, err := tr.Bdl.InvalidateOpenapiOAuth2Token(apistructs.OpenapiOAuth2TokenInvalidateRequest{AccessToken: token})
		if err != nil {
			logrus.Errorf("[alert] reconciler: pipelineID: %d, taskID: %d, task %q failed to invalidate openapi oauth2 token, token: %s, err: %v",
				tr.P.ID, tr.Task.ID, tr.Task.Name, token, err)
		}
	}
}

func (tr *TaskRun) TeardownConcurrencyCount() {
	currentCount := tr.GetTaskConcurrencyCount()
	if currentCount == 0 {
		return
	}
	tr.AddTaskConcurrencyCount(-1)
}

func (tr *TaskRun) TeardownPriorityQueue() {
	popSuccess, popDetail := tr.Throttler.PopProcessing(tr.Task.Extra.UUID)
	if !popSuccess {
		rlog.TWarnf(tr.P.ID, tr.Task.ID, "throttler: pop processing failed, detail: %+v\n", popDetail)
	}
}
