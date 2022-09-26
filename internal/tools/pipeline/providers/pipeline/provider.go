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

package pipeline

import (
	"context"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/actionmgr"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/app"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/cache"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/cancel"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/edgepipeline_register"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/edgereporter"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/permission"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/queuemanager"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/resource"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/run"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/secret"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/user"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
}

// +provider
type provider struct {
	Cfg      *config
	Log      logs.Logger
	Register transport.Register
	MySQL    mysqlxorm.Interface

	*pipelineService

	App          app.Interface
	User         user.Interface
	ActionMgr    actionmgr.Interface
	CronService  cronpb.CronServiceServer `autowired:"erda.core.pipeline.cron.CronService" required:"true"`
	EdgeRegister edgepipeline_register.Interface
	EdgeReporter edgereporter.Interface
	QueueManager queuemanager.Interface
	Resource     resource.Interface
	Secret       secret.Interface
	PipeRun      run.Interface
	Cache        cache.Interface
	Permission   permission.Interface
	Cancel       cancel.Interface
}

func (p *provider) Init(ctx servicehub.Context) error {
	bdl := bundle.New(bundle.WithErdaServer())
	p.pipelineService = &pipelineService{
		p:        p,
		dbClient: &dbclient.Client{Engine: p.MySQL.DB()},
		bdl:      bdl,

		appSvc:       p.App,
		user:         p.User,
		actionMgr:    p.ActionMgr,
		cronSvc:      p.CronService,
		edgeRegister: p.EdgeRegister,
		edgeReporter: p.EdgeReporter,
		queueManage:  p.QueueManager,
		resource:     p.Resource,
		secret:       p.Secret,
		run:          p.PipeRun,
		cache:        p.Cache,
		permission:   p.Permission,
		cancel:       p.Cancel,
	}
	if p.Register != nil {
		pb.RegisterPipelineServiceImp(p.Register, p.pipelineService, apis.Options())
	}
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	return nil
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("erda.core.pipeline.pipeline", &servicehub.Spec{
		Types:                []reflect.Type{interfaceType},
		OptionalDependencies: []string{"service-register"},
		Description:          "pipeline service",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
