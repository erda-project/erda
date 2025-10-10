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
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache"
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
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.registerAIProxyManageAPI()
	p.registerMcpProxyManageAPI()

	// initialize cache manager
	p.SetCacheManager(cache.NewCacheManager(p.Dao, p.L, false))

	p.ServeAIProxyV2()

	return nil
}

func (p *provider) registerAIProxyManageAPI() {
	encoderOpts := mux.InfraEncoderOpt(mux.InfraCORS)
	clientpb.RegisterClientServiceImp(p, &handler_client.ClientHandler{DAO: p.Dao}, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.Dao), permission.CheckClientPerm)
	modelproviderpb.RegisterModelProviderServiceImp(p, &handler_model_provider.ModelProviderHandler{DAO: p.Dao}, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.Dao), permission.CheckModelProviderPerm)
	modelpb.RegisterModelServiceImp(p, &handler_model.ModelHandler{DAO: p.Dao}, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.Dao), permission.CheckModelPerm)
	clientmodelrelationpb.RegisterClientModelRelationServiceImp(p, &handler_client_model_relation.ClientModelRelationHandler{DAO: p.Dao}, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.Dao), permission.CheckClientModelRelationPerm)
	promptpb.RegisterPromptServiceImp(p, &handler_prompt.PromptHandler{DAO: p.Dao}, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.Dao), permission.CheckPromptPerm)
	sessionpb.RegisterSessionServiceImp(p, &handler_session.SessionHandler{DAO: p.Dao}, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.Dao), permission.CheckSessionPerm)
	clienttokenpb.RegisterClientTokenServiceImp(p, &handler_client_token.ClientTokenHandler{DAO: p.Dao}, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.Dao), permission.CheckClientTokenPerm)
	i18npb.RegisterI18NServiceImp(p, &handler_i18n.I18nHandler{DAO: p.Dao}, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.Dao), permission.CheckI18nPerm)
	p.SetRichClientHandler(&handler_rich_client.ClientHandler{DAO: p.Dao})
	richclientpb.RegisterRichClientServiceImp(p, p.GetRichClientHandler(), apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.Dao), permission.CheckRichClientPerm, reverseproxy.TrySetLang())
	auditpb.RegisterAuditServiceImp(p, &handler_audit.AuditHandler{DAO: p.Dao}, apis.Options(), encoderOpts, reverseproxy.TrySetAuth(p.Dao), permission.CheckAuditPerm)
}

func (p *provider) registerMcpProxyManageAPI() {
	// for legacy reason, mcp-list api is provided by ai-proxy, so we need to register it for both ai-proxy and mcp-proxy
	mcppb.RegisterMCPServerServiceImp(p, handler_mcp_server.NewMCPHandler(p.Dao, p.Config.McpProxyPublicURL), apis.Options(), reverseproxy.TrySetAuth(p.Dao), permission.CheckMCPPerm)
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
