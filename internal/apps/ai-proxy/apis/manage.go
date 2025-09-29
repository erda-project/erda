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

package apis

import (
	"net/http"

	"github.com/erda-project/erda-infra/pkg/transport"
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
	"github.com/erda-project/erda/internal/pkg/gorilla/mux"
	"github.com/erda-project/erda/pkg/common/apis"
)

type Combined interface {
	Interface
	transport.Register
}

func RegisterAIProxyManageAPI[T Combined](p T) {
	encoderOpts := mux.InfraEncoderOpt(mux.InfraCORS)
	clientpb.RegisterClientServiceImp(p, &handler_client.ClientHandler{DAO: p.GetDBClient()}, apis.Options(), encoderOpts, trySetAuth(p.GetDBClient()), permission.CheckClientPerm)
	modelproviderpb.RegisterModelProviderServiceImp(p, &handler_model_provider.ModelProviderHandler{DAO: p.GetDBClient()}, apis.Options(), encoderOpts, trySetAuth(p.GetDBClient()), permission.CheckModelProviderPerm)
	modelpb.RegisterModelServiceImp(p, &handler_model.ModelHandler{DAO: p.GetDBClient()}, apis.Options(), encoderOpts, trySetAuth(p.GetDBClient()), permission.CheckModelPerm)
	clientmodelrelationpb.RegisterClientModelRelationServiceImp(p, &handler_client_model_relation.ClientModelRelationHandler{DAO: p.GetDBClient()}, apis.Options(), encoderOpts, trySetAuth(p.GetDBClient()), permission.CheckClientModelRelationPerm)
	promptpb.RegisterPromptServiceImp(p, &handler_prompt.PromptHandler{DAO: p.GetDBClient()}, apis.Options(), encoderOpts, trySetAuth(p.GetDBClient()), permission.CheckPromptPerm)
	sessionpb.RegisterSessionServiceImp(p, &handler_session.SessionHandler{DAO: p.GetDBClient()}, apis.Options(), encoderOpts, trySetAuth(p.GetDBClient()), permission.CheckSessionPerm)
	clienttokenpb.RegisterClientTokenServiceImp(p, &handler_client_token.ClientTokenHandler{DAO: p.GetDBClient()}, apis.Options(), encoderOpts, trySetAuth(p.GetDBClient()), permission.CheckClientTokenPerm)
	i18npb.RegisterI18NServiceImp(p, &handler_i18n.I18nHandler{DAO: p.GetDBClient()}, apis.Options(), encoderOpts, trySetAuth(p.GetDBClient()), permission.CheckI18nPerm)
	p.SetRichClientHandler(&handler_rich_client.ClientHandler{DAO: p.GetDBClient()})
	richclientpb.RegisterRichClientServiceImp(p, p.GetRichClientHandler(), apis.Options(), encoderOpts, trySetAuth(p.GetDBClient()), permission.CheckRichClientPerm, trySetLang())
	auditpb.RegisterAuditServiceImp(p, &handler_audit.AuditHandler{DAO: p.GetDBClient()}, apis.Options(), encoderOpts, trySetAuth(p.GetDBClient()), permission.CheckAuditPerm)
}

func RegisterMcpProxyManageAPI[T Combined](p T, mcpPublicUrl string) {
	// for legacy reason, mcp-list api is provided by ai-proxy, so we need to register it for both ai-proxy and mcp-proxy
	mcppb.RegisterMCPServerServiceImp(p, handler_mcp_server.NewMCPHandler(p.GetDBClient(), mcpPublicUrl), apis.Options(), trySetAuth(p.GetDBClient()), permission.CheckMCPPerm)
}

func ServeAIProxyV2[T Combined](p T, h mux.Mux) {
	// support OPTIONS method
	h.HandlePrefix("/", http.MethodOptions, nil, mux.CORS)
	// `/` handle CORS by itself
	h.HandlePrefix("/", "*", ReverseProxyHandle(p))
}
