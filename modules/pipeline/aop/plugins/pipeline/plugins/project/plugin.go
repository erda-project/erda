package project

import (
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
)

type Plugin struct {
	aoptypes.PipelineBaseTunePoint
}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "project" }
func (p *Plugin) Handle(ctx aoptypes.TuneContext) error {
	// 更新项目活跃时间
	pipeline := ctx.SDK.Pipeline
	projectID, err := strconv.ParseUint(pipeline.Labels[apistructs.LabelProjectID], 10, 64)
	if err == nil {
		if err := ctx.SDK.Bundle.UpdateProjectActiveTime(apistructs.ProjectActiveTimeUpdateRequest{
			ProjectID:  projectID,
			ActiveTime: time.Now(),
		}); err != nil {
			logrus.Errorf("failed to update project active time, pipelineID: %d, projectID: %d, err: %v", pipeline.ID, projectID, err)
		}
	}
	return nil
}
