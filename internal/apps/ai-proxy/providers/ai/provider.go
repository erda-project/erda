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

package ai

import (
	_ "embed"
	"net/http"
	"reflect"

	"google.golang.org/grpc"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/providers/grpcserver"
	dynamic "github.com/erda-project/erda-proto-go/core/openapi/dynamic-register/pb"
	proxyapis "github.com/erda-project/erda/internal/apps/ai-proxy/apis"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_rich_client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define"
	"github.com/erda-project/erda/internal/pkg/gorilla/mux"
)

var (
	_ transport.Register = (*provider)(nil)
)

var (
	name         = "erda.app.ai-proxy"
	providerType = reflect.TypeOf((*provider)(nil))
	spec         = servicehub.Spec{
		Services:    []string{"erda.app.ai-proxy.Server"},
		Summary:     "ai-proxy server",
		Description: "Reverse proxy service between AI vendors and client applications, providing a cut-through for service access",
		ConfigFunc:  func() interface{} { return new(config.Config) },
		Types:       []reflect.Type{providerType},
		Creator:     func() servicehub.Provider { return new(provider) },
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

	// register gRPC and http handler
	proxyapis.RegisterAIProxyManageAPI(p)
	proxyapis.RegisterMcpProxyManageAPI(p, p.Config.McpProxyPublicURL)

	// initialize cache manager
	p.cacheManager = cache.NewCacheManager(p.Dao, p.L, p.IsMcpProxy())

	p.HTTP.Handle("/health", http.MethodGet, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	// reverse proxy to AI provider's server
	proxyapis.ServeAIProxyV2(p, p.HTTP)

	return nil
}

func (p *provider) Add(method, path string, h transhttp.HandlerFunc) {
	p.HTTP.Handle(path, method, http.HandlerFunc(h), mux.SetXRequestId, mux.CORS)
}

func (p *provider) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	p.GRPC.RegisterService(desc, impl)
}

func (p *provider) FindBestMatch(method, path string) router_define.RouteMatcher {
	return p.Router.FindBestMatch(method, path)
}

func (p *provider) GetDBClient() dao.DAO {
	return p.Dao
}

func (p *provider) GetRichClientHandler() *handler_rich_client.ClientHandler {
	return p.richClientHandler
}

func (p *provider) SetRichClientHandler(handler *handler_rich_client.ClientHandler) {
	p.richClientHandler = handler
}

func (p *provider) GetCacheManager() cachetypes.Manager {
	return p.cacheManager
}

func (p *provider) IsMcpProxy() bool {
	return false
}
