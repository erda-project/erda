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
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors"
)

// Interface .
type Interface interface {
	transhttp.Router
}

type config struct {
	// Timeout time.Duration `file:"timeout" default:"2m"` // TODO
}

// +provider
type provider struct {
	Cfg          *config
	Log          logs.Logger
	HTTP         httpserver.Router `autowired:"http-server"`
	interceptors interceptors.Interceptors
}

func (p *provider) Init(ctx servicehub.Context) error {
	var inters interceptors.Interceptors
	ctx.Hub().ForeachServices(func(service string) bool {
		if strings.HasPrefix(service, "openapi-interceptor-") {
			inter, ok := ctx.Service(service).(interceptors.Interface)
			if !ok {
				panic(fmt.Errorf("service %s is not interceptor", service))
			}
			inters = append(inters, inter.List()...)
		}
		return true
	})
	sort.Sort(inters)
	p.interceptors = inters
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
		handler = transhttp.HandlerFunc(s.p.interceptors[i].Wrapper(http.HandlerFunc(handler)))
	}
	s.router.Add(method, path, handler, httpserver.WithPathFormat(httpserver.PathFormatGoogleAPIs))
}

func init() {
	servicehub.Register("openapi-ng", &servicehub.Spec{
		Services: []string{"openapi-ng"},
		DependenciesFunc: func(hub *servicehub.Hub) (list []string) {
			hub.ForeachServices(func(service string) bool {
				if strings.HasPrefix(service, "openapi-interceptor-") {
					list = append(list, service)
				}
				return true
			})
			return list
		},
		Types:      []reflect.Type{reflect.TypeOf((*transhttp.Router)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
