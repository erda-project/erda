package echo

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
)

type Plugin struct {
	aoptypes.PipelineBaseTunePoint
}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "echo" }
func (p *Plugin) Handle(ctx aoptypes.TuneContext) error {
	logrus.Debugf("say hello to pipeline AOP, type: %s, trigger: %s, pipelineID: %d, status: %s",
		ctx.SDK.TuneType, ctx.SDK.TuneTrigger, ctx.SDK.Pipeline.ID, ctx.SDK.Pipeline.Status)
	return nil
}
