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
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_audit"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_client_model_relation"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_i18n"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_model"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_model_provider"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_prompt"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_rich_client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_session"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_token_usage"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/ai-proxy/aiproxytypes"
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
