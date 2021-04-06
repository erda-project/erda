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
