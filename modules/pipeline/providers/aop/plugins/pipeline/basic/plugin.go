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

package basic

import (
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/plugins_manage"
)

type Plugin struct {
	aoptypes.PipelineBaseTunePoint
}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "basic" }
func (p *Plugin) Handle(ctx *aoptypes.TuneContext) error {
	pipeline := ctx.SDK.Pipeline

	// make report content
	var content apistructs.PipelineBasicReport
	content.PipelineSource = pipeline.PipelineSource
	content.PipelineYmlName = pipeline.PipelineYmlName
	content.ClusterName = pipeline.ClusterName
	content.TimeCreated = pipeline.TimeCreated
	content.TimeBegin = pipeline.TimeBegin
	content.TimeEnd = pipeline.TimeEnd
	content.TotalCostTimeSec = pipeline.CostTimeSec

	tasks, err := ctx.SDK.DBClient.ListPipelineTasksByPipelineID(pipeline.ID)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		taskReport := apistructs.TaskReportInfo{
			Name:             task.Name,
			ActionType:       task.Type,
			ActionVersion:    task.Extra.Action.Version,
			ExecutorType:     string(task.ExecutorKind),
			ClusterName:      task.Extra.ClusterName,
			TimeBegin:        getTimeOrNil(task.TimeBegin),
			TimeEnd:          getTimeOrNil(task.TimeEnd),
			TimeBeginQueue:   getTimeOrNil(task.Extra.TimeBeginQueue),
			TimeEndQueue:     getTimeOrNil(task.Extra.TimeEndQueue),
			QueueCostTimeSec: task.QueueTimeSec,
			RunCostTimeSec:   task.CostTimeSec,
			MachineStat:      task.Result.MachineStat,
			Meta: func() map[string]string {
				result := make(map[string]string)
				for _, meta := range task.Result.Metadata {
					result[meta.Name] = meta.Value
				}
				return result
			}(),
		}
		content.TaskInfos = append(content.TaskInfos, taskReport)
	}

	_, err = ctx.SDK.Report.Create(apistructs.PipelineReportCreateRequest{
		PipelineID: ctx.SDK.Pipeline.ID,
		Type:       apistructs.PipelineReportTypeBasic,
		Meta:       apistructs.PipelineReportMeta{"data": content},
	})
	return err
}

func getTimeOrNil(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

type config struct {
	TuneType    aoptypes.TuneType      `file:"tune_type"`
	TuneTrigger []aoptypes.TuneTrigger `file:"tune_trigger" `
}

// +provider
type provider struct {
	Cfg *config
}

func (p *provider) Init(ctx servicehub.Context) error {
	for _, tuneTrigger := range p.Cfg.TuneTrigger {
		err := plugins_manage.RegisterTunePointToTuneGroup(p.Cfg.TuneType, tuneTrigger, New())
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func init() {
	servicehub.Register("erda.core.pipeline.aop.plugins.pipeline.basic", &servicehub.Spec{
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
