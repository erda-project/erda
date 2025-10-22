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
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
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
	tokenusagepb "github.com/erda-project/erda-proto-go/apps/aiproxy/usage/token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/usage/token_usage"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_mcp_server"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/permission"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/ai-proxy/aiproxytypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/reverseproxy"
	"github.com/erda-project/erda/internal/pkg/gorilla/mux"
	"github.com/erda-project/erda/pkg/common/apis"
)

const Name = "erda.app.ai-proxy"

type Config struct {
	McpProxyPublicURL string `file:"mcp_proxy_public_url" env:"MCP_PROXY_PUBLIC_URL"`
}

type provider struct {
	Config *Config
	L      logs.Logger
	Dao    dao.DAO `autowired:"erda.apps.ai-proxy.dao"`

	reverseproxy.Interface `autowired:"erda.app.reverse-proxy"`

	cache cachetypes.Manager

	handlers           *aiproxytypes.Handlers
	ctxhelperFunctions []func(context.Context)
}

func (p *provider) Init(ctx servicehub.Context) error {

	// initialize cache manager
	p.cache = cache.NewCacheManager(p.Dao, p.L, false)
	p.SetCacheManager(p.cache)

	// initialize token usage collector
	token_usage.InitUsageCollector(p.Dao)

	p.initHandlers()

	p.registerAIProxyManageAPI()
	p.registerMcpProxyManageAPI()

	p.ServeReverseProxyV2(reverseproxy.WithCtxHelperItems(
		func(ctx context.Context) {
			ctxhelper.PutAIProxyHandlers(ctx, p.handlers)
		},
	))

	return nil
}

func (p *provider) registerAIProxyManageAPI() {
	encoderOpts := mux.InfraEncoderOpt(mux.InfraCORS)
	register := p.Interface

	clientpb.RegisterClientServiceImp(register, p.handlers.ClientHandler, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.cache), permission.CheckClientPerm)
	modelproviderpb.RegisterModelProviderServiceImp(register, p.handlers.ModelProviderHandler, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.cache), permission.CheckModelProviderPerm)
	modelpb.RegisterModelServiceImp(register, p.handlers.ModelHandler, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.cache), permission.CheckModelPerm)
	clientmodelrelationpb.RegisterClientModelRelationServiceImp(register, p.handlers.ClientModelRelationHandler, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.cache), permission.CheckClientModelRelationPerm)
	promptpb.RegisterPromptServiceImp(register, p.handlers.PromptHandler, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.cache), permission.CheckPromptPerm)
	sessionpb.RegisterSessionServiceImp(register, p.handlers.SessionHandler, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.cache), permission.CheckSessionPerm)
	clienttokenpb.RegisterClientTokenServiceImp(register, p.handlers.ClientTokenHandler, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.cache), permission.CheckClientTokenPerm)
	i18npb.RegisterI18NServiceImp(register, p.handlers.I18nHandler, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.cache), permission.CheckI18nPerm)
	richclientpb.RegisterRichClientServiceImp(register, p.handlers.RichClientHandler, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.cache), permission.CheckRichClientPerm, reverseproxy.TrySetLang())
	auditpb.RegisterAuditServiceImp(register, p.handlers.AuditHandler, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.cache), permission.CheckAuditPerm)
	tokenusagepb.RegisterTokenUsageServiceImp(register, p.handlers.TokenUsageHandler, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.cache), permission.CheckTokenUsagePerm, reverseproxy.TrySetLang())
}

func (p *provider) registerMcpProxyManageAPI() {
	// for legacy reason, mcp-list api is provided by ai-proxy, so we need to register it for both ai-proxy and mcp-proxy
	mcppb.RegisterMCPServerServiceImp(p, handler_mcp_server.NewMCPHandler(p.Dao, p.Config.McpProxyPublicURL), apis.Options(), reverseproxy.TrySetAuth(p.cache), permission.CheckMCPPerm)
}

func init() {
	servicehub.Register(Name, &servicehub.Spec{
		Services:    []string{"erda.app.ai-proxy.Server"},
		Summary:     "ai-proxy server",
		Description: "Reverse proxy service between AI vendors and client applications, providing a cut-through for service access",
		ConfigFunc:  func() interface{} { return new(Config) },
		Types:       []reflect.Type{reflect.TypeOf((*provider)(nil))},
		Creator:     func() servicehub.Provider { return new(provider) },
	})
}
