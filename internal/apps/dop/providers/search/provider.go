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

package search

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/dop/search/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct{}

type provider struct {
	bdl      *bundle.Bundle
	Cfg      *config
	Log      logs.Logger
	Register transport.Register

	Query         query.Interface
	SearchService *ServiceImpl
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.dop.search.SearchService" || ctx.Type() == pb.SearchServiceServerType() || ctx.Type() == pb.SearchServiceHandlerType():
		return p.SearchService
	}
	return p
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithErdaServer())
	p.SearchService = &ServiceImpl{
		query: p.Query,
		bdl:   p.bdl,
	}
	if p.Register != nil {
		pb.RegisterSearchServiceImp(p.Register, p.SearchService, apis.Options())
	}
	return nil
}

func init() {
	servicehub.Register("erda.dop.search", &servicehub.Spec{
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
