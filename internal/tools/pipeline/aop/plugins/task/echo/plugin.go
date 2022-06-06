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

package echo

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/tools/pipeline/aop"
	aoptypes2 "github.com/erda-project/erda/modules/tools/pipeline/aop/aoptypes"
)

// +provider
type provider struct {
	aoptypes2.TaskBaseTunePoint
}

func (p *provider) Name() string { return "echo" }

func (p *provider) Handle(ctx *aoptypes2.TuneContext) error {
	logrus.Debugf("say hello to task AOP, type: %s, trigger: %s, pipelineID: %d, taskID: %d, status: %s",
		ctx.SDK.TuneType, ctx.SDK.TuneTrigger, ctx.SDK.Pipeline.ID, ctx.SDK.Task.ID, ctx.SDK.Task.Status)
	return nil
}

func (p *provider) Init(ctx servicehub.Context) error {
	err := aop.RegisterTunePoint(p)
	if err != nil {
		panic(err)
	}
	return nil
}

func init() {
	servicehub.Register(aop.NewProviderNameByPluginName(&provider{}), &servicehub.Spec{
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
