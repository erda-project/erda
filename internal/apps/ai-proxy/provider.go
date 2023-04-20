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

package ai_proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	_ "github.com/erda-project/erda-infra/providers/health"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/filter"
	provider2 "github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	route2 "github.com/erda-project/erda/internal/pkg/ai-proxy/route"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	name         = "erda.app.ai-proxy"
	providerType = reflect.TypeOf((*provider)(nil))
	spec         = servicehub.Spec{
		Services:    []string{"erda.app.ai-proxy.Server"},
		Summary:     "ai-proxy server",
		Description: "Reverse proxy service between AI vendors and client applications, providing a cut-through for service access",
		ConfigFunc: func() interface{} {
			return new(config)
		},
		Types: []reflect.Type{providerType},
		Creator: func() servicehub.Provider {
			return new(provider)
		},
	}
)

func init() {
	servicehub.Register(name, &spec)
}

type provider struct {
	L          logs.Logger
	Config     *config
	HttpServer httpserver.Router `autowired:"http-server"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.L.Info("providers config:\n%s", strutil.TryGetYamlStr(p.Config.Providers))
	p.L.Info("routes config:\n%s", strutil.TryGetYamlStr(p.Config.Routes))
	p.HttpServer.Any("/", p.ServeHTTP)
	return nil
}

func (p *provider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rout, ok := p.matchRoute(r.URL.Path, r.Method)
	if !ok {
		p.responseNoSuchRoute(w, r.URL.Path)
		return
	}
	var ctx = p.ctxWith(rout, p.Config.Providers)
	for i := 0; i < len(rout.Filters); i++ {
		if signal := rout.Filters[i].OnHttpRequest(ctx, w, r); signal != filter.Continue {
			return
		}
	}
	for i := len(rout.Filters) - 1; i >= 0; i-- {
		if signal := rout.Filters[i].OnHttpResponse(ctx, w, r); signal != filter.Continue {
			return
		}
	}
}

func (p *provider) matchRoute(path, method string) (*route2.Route, bool) {
	// todo: 应当改成树形数据结构来存储和查找 route, 不过在 route 数量有限的情形下影响不大
	for _, r := range p.Config.Routes {
		if r.Match(path, method) {
			return r, true
		}
	}
	return nil, false
}

func (p *provider) responseNoSuchRoute(w http.ResponseWriter, path string) {
	w.Header().Set("server", "ai-proxy/erda")
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": "no such route",
		"path":  path,
	})
}

func (p *provider) ctxWith(route *route2.Route, providers provider2.Providers) context.Context {
	return context.WithValue(context.WithValue(context.Background(), filter.RouteCtxKey{}, route), filter.ProviderCtxKey{}, providers)
}

type config struct {
	HttpServer struct {
		Addr string
	} `json:"httpServer" yaml:"httpServer"`
	Providers provider2.Providers `json:"providers" yaml:"providers"`
	Routes    route2.Routes       `json:"routes" yaml:"routes"`
}
