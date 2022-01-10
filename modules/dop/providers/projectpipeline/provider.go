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

package projectpipeline

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/dop/projectpipeline/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type config struct {
}

type provider struct {
	Cfg                *config
	Log                logs.Logger
	bundle             *bundle.Bundle
	DB                 *gorm.DB           `autowired:"mysql-client"`
	Register           transport.Register `autowired:"service-register" required:"true"`

	projectPipelineSvc *ProjectPipelineService
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bundle = bundle.New(bundle.WithCoreServices())
	p.projectPipelineSvc = &ProjectPipelineService{
		logger: p.Log,
		bundle: p.bundle,
		db: &dao.DBClient{
			DBEngine: &dbengine.DBEngine{
				DB: p.DB,
			},
		},
	}
	if p.Register != nil {
		pb.RegisterProjectPipelineServiceImp(p.Register, p.projectPipelineSvc, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.dop.projectpipeline.ProjectPipelineService" || ctx.Type() == pb.ProjectPipelineServiceServerType() || ctx.Type() == pb.ProjectPipelineServiceHandlerType():
		return p.projectPipelineSvc
	}
	return p
}

func init() {
	servicehub.Register("erda.dop.projectpipeline", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		Description:          "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
