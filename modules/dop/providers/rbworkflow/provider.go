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

package rbworkflow

import (
	"context"
	"reflect"

	"github.com/erda-project/erda/pkg/common/apis"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/dop/rbworkflow/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/providers/rbworkflow/db"
)

type config struct {
}

type provider struct {
	Cfg      *config
	Log      logs.Logger
	bundle   *bundle.Bundle
	DB       *gorm.DB           `autowired:"mysql-gorm.v2-client"`
	Register transport.Register `autowired:"service-register" required:"true"`
	Trans    i18n.Translator    `translator:"project-pipeline" required:"true"`

	RbWorkflowSvc *ServiceImplement
}

func (p *provider) Run(ctx context.Context) error {
	return nil
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bundle = bundle.New(bundle.WithGittar())
	p.RbWorkflowSvc = &ServiceImplement{
		db:  &db.Client{DB: p.DB},
		bdl: p.bundle,
	}
	if p.Register != nil {
		pb.RegisterRbWorkflowServiceImp(p.Register, p.RbWorkflowSvc, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.dop.rbWorkflow.RbWorkflowServiceMethod" || ctx.Type() == reflect.TypeOf(reflect.TypeOf((*Service)(nil)).Elem()):
		return p.RbWorkflowSvc
	case ctx.Service() == "erda.dop.rbWorkflow.RbWorkflowService" || ctx.Type() == pb.RbWorkflowServiceServerType() || ctx.Type() == pb.RbWorkflowServiceHandlerType():
		return p.RbWorkflowSvc
	}
	return p
}

func init() {
	servicehub.Register("erda.dop.rbWorkflow", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                append(pb.Types()),
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
