package echo

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
)

type Plugin struct {
	aoptypes.TaskBaseTunePoint
}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "echo" }
func (p *Plugin) Handle(ctx aoptypes.TuneContext) error {
	logrus.Debugf("say hello to task AOP, type: %s, trigger: %s, pipelineID: %d, taskID: %d, status: %s",
		ctx.SDK.TuneType, ctx.SDK.TuneTrigger, ctx.SDK.Pipeline.ID, ctx.SDK.Task.ID, ctx.SDK.Task.Status)
	return nil
}
