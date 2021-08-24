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

package echo

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/plugins_manage"
)

type Plugin struct {
	aoptypes.PipelineBaseTunePoint
}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "echo" }
func (p *Plugin) Handle(ctx *aoptypes.TuneContext) error {
	logrus.Debugf("say hello to pipeline AOP, type: %s, trigger: %s, pipelineID: %d, status: %s",
		ctx.SDK.TuneType, ctx.SDK.TuneTrigger, ctx.SDK.Pipeline.ID, ctx.SDK.Pipeline.Status)
	return nil
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
	servicehub.Register("erda.core.pipeline.aop.plugins.pipeline.echo", &servicehub.Spec{
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
