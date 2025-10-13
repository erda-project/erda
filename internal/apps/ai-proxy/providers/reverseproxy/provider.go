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

package reverseproxy

import (
	"context"
	_ "embed"
	"net/http"
	"reflect"

	"google.golang.org/grpc"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	"github.com/erda-project/erda-infra/providers/grpcserver"
	dynamic "github.com/erda-project/erda-proto-go/core/openapi/dynamic-register/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth/akutil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_rich_client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define"
	"github.com/erda-project/erda/internal/pkg/gorilla/mux"
	"github.com/erda-project/erda/pkg/common/apis"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

type Interface interface {
	transport.Register
	SetCacheManager(cachetypes.Manager)
}

var (
	_ transport.Register = (*provider)(nil)
)

var (
	name          = "erda.app.reverse-proxy"
	interfaceType = reflect.TypeOf((*Interface)(nil)).Elem()
	spec          = servicehub.Spec{
		Services:    []string{"erda.app.reverse-proxy"},
		Summary:     "reverse-proxy server",
		Description: "Reverse proxy service framework",
		ConfigFunc:  func() interface{} { return new(config.Config) },
		Types:       []reflect.Type{interfaceType},
		Creator:     func() servicehub.Provider { return new(provider) },
	}
	TrySetAuth = func(dao dao.DAO) transport.ServiceOption {
		return transport.WithInterceptors(func(h interceptor.Handler) interceptor.Handler {
			return func(ctx context.Context, req interface{}) (interface{}, error) {
				ctx = ctxhelper.InitCtxMapIfNeed(ctx)
				// check admin key first
				isAdmin, err := akutil.CheckAdmin(ctx, req, dao)
				if err != nil {
					return nil, err
				}
				if isAdmin {
					ctxhelper.PutIsAdmin(ctx, true)
					return h(ctx, req)
				}
				// try set clientId by ak
				clientToken, client, err := akutil.CheckAkOrToken(ctx, req, dao)
				if err != nil {
					return nil, err
				}
				if clientToken != nil {
					ctxhelper.PutClientToken(ctx, clientToken)
				}
				if client != nil {
					ctxhelper.PutClient(ctx, client)
					ctxhelper.PutClientId(ctx, client.Id)
				}
				return h(ctx, req)
			}
		})
	}
	TrySetLang = func() transport.ServiceOption {
		return transport.WithInterceptors(func(h interceptor.Handler) interceptor.Handler {
			return func(ctx context.Context, req interface{}) (interface{}, error) {
				ctx = ctxhelper.InitCtxMapIfNeed(ctx)
				lang := apis.GetHeader(ctx, httperrorutil.HeaderKeyAcceptLanguage)
				if len(lang) > 0 {
					ctxhelper.PutAccessLang(ctx, lang)
				}
				return h(ctx, req)
			}
		})
	}
)

func init() {
	servicehub.Register(name, &spec)
}

type provider struct {
	Config         *config.Config
	L              logs.Logger
	HTTP           mux.Mux                              `autowired:"gorilla-mux@ai"`
	GRPC           grpcserver.Interface                 `autowired:"grpc-server@ai"`
	Dao            dao.DAO                              `autowired:"erda.apps.ai-proxy.dao"`
	DynamicOpenapi dynamic.DynamicOpenapiRegisterServer `autowired:"erda.core.openapi.dynamic_register.DynamicOpenapiRegister"`

	richClientHandler *handler_rich_client.ClientHandler

	cacheManager cachetypes.Manager

	Routes []*router_define.Route
	Router *router_define.Router
}

func (p *provider) Init(ctx servicehub.Context) error {
	// config
	if err := p.Config.DoPost(); err != nil {
		return err
	}

	// load route configs
	yamlFile, err := router_define.LoadRoutesFromEmbeddedDir(p.Config.EmbedRoutesFS)
	if err != nil {
		return err
	}
	// create rout
	p.Routes = yamlFile.Routes
	p.Router = router_define.NewRouter()
	for _, route := range p.Routes {
		expandedRoutes, err := router_define.ExpandRoute(route)
		if err != nil {
			return err
		}
		for _, expandedRoute := range expandedRoutes {
			p.Router.AddRoute(expandedRoute)
			p.L.Infof("register route from yaml: %s", expandedRoute)
		}
	}

	p.HTTP.Handle("/health", http.MethodGet, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	p.ServeReverseProxyV2()
	return nil
}

func (p *provider) ServeReverseProxyV2() {
	// support OPTIONS method
	p.HTTP.HandlePrefix("/", http.MethodOptions, nil, mux.CORS)
	// fallback to reverse proxy when no route matches
	fallback := p.HandleReverseProxyAPI()
	p.HTTP.HandleNotFound(fallback)
	p.HTTP.HandleMethodNotAllowed(fallback)
}

func (p *provider) SetCacheManager(manager cachetypes.Manager) {
	p.cacheManager = manager
}

func (p *provider) Add(method, path string, h transhttp.HandlerFunc) {
	p.HTTP.Handle(path, method, http.HandlerFunc(h), mux.SetXRequestId, mux.CORS)
}

func (p *provider) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	p.GRPC.RegisterService(desc, impl)
}
