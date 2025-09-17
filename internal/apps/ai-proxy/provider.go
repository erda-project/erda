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
	auditpb "github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	richclientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/rich_client/pb"
	clientmodelrelationpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	i18npb "github.com/erda-project/erda-proto-go/apps/aiproxy/i18n/pb"
	mcppb "github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	promptpb "github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	dynamic "github.com/erda-project/erda-proto-go/core/openapi/dynamic-register/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/common/akutil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_audit"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_client_model_relation"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_i18n"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_mcp_server"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_model"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_model_provider"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_prompt"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_rich_client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_session"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/permission"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define"
	"github.com/erda-project/erda/internal/pkg/gorilla/mux"
	"github.com/erda-project/erda/pkg/common/apis"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
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
	trySetAuth = func(dao dao.DAO) transport.ServiceOption {
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
	trySetLang = func() transport.ServiceOption {
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

	// register gRPC and http handler
	p.registerAIProxyManageAPI()
	p.registerMcpProxyManageAPI()

	// initialize cache manager
	p.cacheManager = cache.NewCacheManager(p.Dao, p.L, p.Config.IsMcpProxy)

	p.HTTP.Handle("/health", http.MethodGet, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	// reverse proxy to AI provider's server
	p.ServeAIProxyV2()

	return nil
}

func (p *provider) ServeAIProxyV2() {
	// support OPTIONS method
	p.HTTP.HandlePrefix("/", http.MethodOptions, nil, mux.CORS)
	// `/` handle CORS by itself
	p.HTTP.HandlePrefix("/", "*", p.HandleReverseProxyAPI())
}

func (p *provider) Add(method, path string, h transhttp.HandlerFunc) {
	p.HTTP.Handle(path, method, http.HandlerFunc(h), mux.SetXRequestId, mux.CORS)
}

func (p *provider) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	p.GRPC.RegisterService(desc, impl)
}

func (p *provider) registerAIProxyManageAPI() {
	if p.Config.IsMcpProxy {
		return
	}
	encoderOpts := mux.InfraEncoderOpt(mux.InfraCORS)
	clientpb.RegisterClientServiceImp(p, &handler_client.ClientHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckClientPerm)
	modelproviderpb.RegisterModelProviderServiceImp(p, &handler_model_provider.ModelProviderHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckModelProviderPerm)
	modelpb.RegisterModelServiceImp(p, &handler_model.ModelHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckModelPerm)
	clientmodelrelationpb.RegisterClientModelRelationServiceImp(p, &handler_client_model_relation.ClientModelRelationHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckClientModelRelationPerm)
	promptpb.RegisterPromptServiceImp(p, &handler_prompt.PromptHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckPromptPerm)
	sessionpb.RegisterSessionServiceImp(p, &handler_session.SessionHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckSessionPerm)
	clienttokenpb.RegisterClientTokenServiceImp(p, &handler_client_token.ClientTokenHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckClientTokenPerm)
	i18npb.RegisterI18NServiceImp(p, &handler_i18n.I18nHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckI18nPerm)
	p.richClientHandler = &handler_rich_client.ClientHandler{DAO: p.Dao}
	richclientpb.RegisterRichClientServiceImp(p, p.richClientHandler, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckRichClientPerm, trySetLang())
	auditpb.RegisterAuditServiceImp(p, &handler_audit.AuditHandler{DAO: p.Dao}, apis.Options(), encoderOpts, trySetAuth(p.Dao), permission.CheckAuditPerm)
}

func (p *provider) registerMcpProxyManageAPI() {
	// for legacy reason, mcp-list api is provided by ai-proxy, so we need to register it for both ai-proxy and mcp-proxy
	mcppb.RegisterMCPServerServiceImp(p, handler_mcp_server.NewMCPHandler(p.Dao, p.Config.McpProxyPublicURL), apis.Options(), trySetAuth(p.Dao), permission.CheckMCPPerm)
}
