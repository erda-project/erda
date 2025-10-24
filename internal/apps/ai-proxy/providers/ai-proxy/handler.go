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

	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
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
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
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
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_token_usage"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/permission"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/ai-proxy/aiproxytypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/reverseproxy"
	"github.com/erda-project/erda/internal/pkg/gorilla/mux"
	"github.com/erda-project/erda/pkg/common/apis"
)

func (p *provider) initHandlers() {
	p.handlers = &aiproxytypes.Handlers{
		ClientHandler:              &handler_client.ClientHandler{DAO: p.Dao},
		ModelHandler:               &handler_model.ModelHandler{DAO: p.Dao},
		ModelProviderHandler:       &handler_model_provider.ModelProviderHandler{DAO: p.Dao},
		ClientModelRelationHandler: &handler_client_model_relation.ClientModelRelationHandler{DAO: p.Dao},
		PromptHandler:              &handler_prompt.PromptHandler{DAO: p.Dao},
		SessionHandler:             &handler_session.SessionHandler{DAO: p.Dao},
		ClientTokenHandler:         &handler_client_token.ClientTokenHandler{DAO: p.Dao},
		I18nHandler:                &handler_i18n.I18nHandler{DAO: p.Dao},
		RichClientHandler:          &handler_rich_client.ClientHandler{DAO: p.Dao},
		AuditHandler:               &handler_audit.AuditHandler{DAO: p.Dao},
		TokenUsageHandler:          &handler_token_usage.TokenUsageHandler{DAO: p.Dao, Cache: p.cache},
	}
}

func (p *provider) registerAIProxyManageAPI() {
	register := p.ReverseProxy

	clientpb.RegisterClientServiceImp(register, p.handlers.ClientHandler, p.getProtoOptions(permission.CheckClientPerm)...)
	modelproviderpb.RegisterModelProviderServiceImp(register, p.handlers.ModelProviderHandler, p.getProtoOptions(permission.CheckModelProviderPerm)...)
	modelpb.RegisterModelServiceImp(register, p.handlers.ModelHandler, p.getProtoOptions(permission.CheckModelPerm)...)
	clientmodelrelationpb.RegisterClientModelRelationServiceImp(register, p.handlers.ClientModelRelationHandler, p.getProtoOptions(permission.CheckClientModelRelationPerm)...)
	promptpb.RegisterPromptServiceImp(register, p.handlers.PromptHandler, p.getProtoOptions(permission.CheckPromptPerm)...)
	sessionpb.RegisterSessionServiceImp(register, p.handlers.SessionHandler, p.getProtoOptions(permission.CheckSessionPerm)...)
	clienttokenpb.RegisterClientTokenServiceImp(register, p.handlers.ClientTokenHandler, p.getProtoOptions(permission.CheckClientTokenPerm)...)
	i18npb.RegisterI18NServiceImp(register, p.handlers.I18nHandler, p.getProtoOptions(permission.CheckI18nPerm)...)
	richclientpb.RegisterRichClientServiceImp(register, p.handlers.RichClientHandler, p.getProtoOptions(permission.CheckRichClientPerm)...)
	auditpb.RegisterAuditServiceImp(register, p.handlers.AuditHandler, p.getProtoOptions(permission.CheckAuditPerm)...)
	tokenusagepb.RegisterTokenUsageServiceImp(register, p.handlers.TokenUsageHandler, p.getProtoOptions(permission.CheckTokenUsagePerm)...)
}

func (p *provider) registerMcpProxyManageAPI() {
	// for legacy reason, mcp-list api is provided by ai-proxy, so we need to register it for both ai-proxy and mcp-proxy
	mcppb.RegisterMCPServerServiceImp(p.ReverseProxy, handler_mcp_server.NewMCPHandler(p.Dao, p.Config.McpProxyPublicURL), apis.Options(), reverseproxy.TrySetAuth(p.cache), permission.CheckMCPPerm)
}

func (p *provider) getProtoOptions(opts ...transport.ServiceOption) []transport.ServiceOption {
	unifiedOpts := []transport.ServiceOption{
		apis.Options(),
		mux.InfraEncoderOpt(mux.InfraCORS),
		reverseproxy.TrySetAuth(p.cache),
		reverseproxy.TrySetLang(),
		setContextMap(p),
	}
	return append(unifiedOpts, opts...)
}

var setContextMap = func(p *provider) transport.ServiceOption {
	return transport.WithInterceptors(func(h interceptor.Handler) interceptor.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx = ctxhelper.InitCtxMapIfNeed(ctx)
			ctxhelper.PutDBClient(ctx, p.Dao)
			ctxhelper.PutCacheManager(ctx, p.cache)
			return h(ctx, req)
		}
	})
}
