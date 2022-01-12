// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package definitionreport

import (
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda/modules/pipeline/aop"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/providers/definition/db"
)

// +provider
type provider struct {
	MySQL mysqlxorm.Interface `autowired:"mysql-xorm"`
	aoptypes.PipelineBaseTunePoint
	client *db.Client
}

func (p *provider) Name() string { return "definition-report" }

func (p *provider) Handle(ctx *aoptypes.TuneContext) error {
	pipeline := ctx.SDK.Pipeline
	if pipeline.PipelineDefinitionId == "" {
		return nil
	}
	if !pipeline.Status.IsEndStatus() {
		return nil
	}
	definition, err := p.client.GetPipelineDefinition(pipeline.PipelineDefinitionId)
	if err != nil {
		return err
	}

	definition.Executor = definition.Creator
	definition.CostTime = uint64(pipeline.CostTimeSec)
	definition.StartedAt = *pipeline.TimeBegin
	definition.EndedAt = *pipeline.TimeEnd
	definition.PipelineID = pipeline.ID

	err = p.client.UpdatePipelineDefinition(definition.ID, definition)
	if err != nil {
		return err
	}
	return nil
}

func (p *provider) Init(ctx servicehub.Context) error {
	err := aop.RegisterTunePoint(p)
	if err != nil {
		panic(err)
	}
	p.client = &db.Client{Interface: p.MySQL}
	return nil
}

func init() {
	servicehub.Register(aop.NewProviderNameByPluginName(&provider{}), &servicehub.Spec{
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
