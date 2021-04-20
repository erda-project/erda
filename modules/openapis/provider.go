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

package openapis

import (
	"net/http"
	"reflect"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/providers/httpserver"

	// http interceptors
	auth "github.com/erda-project/erda/modules/openapis/interceptors/auth"
	csrf "github.com/erda-project/erda/modules/openapis/interceptors/csrf"
)

// Interface .
type Interface interface {
	transhttp.Router
}

type config struct {
	Timeout time.Duration `file:"timeout" default:"2m"`
}

type Interceptor func(h http.HandlerFunc) http.HandlerFunc

// +provider
type provider struct {
	Cfg          *config
	Log          logs.Logger
	HTTP         httpserver.Router `autowired:"http-server"`
	interceptors []Interceptor
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.interceptors = append(p.interceptors,
		csrf.Interceptor,
		auth.Interceptor,
	)
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return &service{
		p:      p,
		name:   ctx.Caller(),
		router: p.HTTP,
	}
}

var _ Interface = (*service)(nil)

type service struct {
	p      *provider
	name   string
	router httpserver.Router
}

func (s *service) Add(method, path string, handler transhttp.HandlerFunc) {
	for i := len(s.p.interceptors) - 1; i >= 0; i-- {
		handler = transhttp.HandlerFunc(s.p.interceptors[i](http.HandlerFunc(handler)))
	}
	s.router.Add(method, path, handler, httpserver.WithPathFormat(httpserver.PathFormatGoogleAPIs))
}

func init() {
	servicehub.Register("openapis", &servicehub.Spec{
		Services: []string{"openapis"},
		Types:    []reflect.Type{reflect.TypeOf((*transhttp.Router)(nil)).Elem()},
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
