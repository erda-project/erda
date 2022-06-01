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

package build

import (
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda-proto-go/core/pipeline/build/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/providers/build/db"
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

	bdl          *bundle.Bundle
	dbClient     *db.Client
	buildService *buildService
}

func (p *provider) Init(ctx servicehub.Context) error {
	dbClient := &db.Client{Interface: p.MySQL}
	p.buildService = &buildService{p: p, dbClient: dbClient}
	p.dbClient = dbClient
	if p.Register != nil {
		pb.RegisterBuildServiceImp(p.Register, p.buildService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.pipeline.build.BuildService" || ctx.Type() == pb.BuildServiceServerType() || ctx.Type() == pb.BuildServiceHandlerType():
		return p.buildService
	}
	return p
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("erda.core.pipeline.build", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                append([]reflect.Type{interfaceType}, pb.Types()...),
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
