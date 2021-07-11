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

package openapi

import (
	"context"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/openapi-ng/api"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors"
	discover "github.com/erda-project/erda/providers/service-discover"
	"github.com/recallsong/go-utils/errorx"
)

// Interface .
type Interface interface {
	transhttp.Router
	AddAPI(spec *api.Spec)
}

type config struct{}

// +provider
type provider struct {
	Cfg          *config
	Log          logs.Logger
	HTTP         httpserver.Router  `autowired:"http-server"`
	Discover     discover.Interface `autowired:"discover"`
	interceptors interceptors.Interceptors
	errors       []error
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.interceptors = interceptors.GetInterceptors(ctx)
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	if len(p.errors) > 0 {
		return errorx.Errors(p.errors)
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return &router{
		name:         ctx.Caller(),
		log:          p.Log,
		discover:     p.Discover,
		interceptors: p.interceptors,
		http:         p.HTTP,
		addError: func(err error) {
			p.errors = append(p.errors, err)
		},
	}
}

func init() {
	servicehub.Register("openapi-ng", &servicehub.Spec{
		Services:         []string{"openapi-ng"},
		DependenciesFunc: interceptors.GetInterceptorServices,
		Types:            []reflect.Type{reflect.TypeOf((*transhttp.Router)(nil)).Elem()},
		ConfigFunc:       func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
