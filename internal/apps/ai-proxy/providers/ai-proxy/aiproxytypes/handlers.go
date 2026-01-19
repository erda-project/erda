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

package aiproxytypes

import (
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_audit"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_cache"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_client_model_relation"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_i18n"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_model"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_prompt"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_rich_client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_service_provider"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_session"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_template"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_token_usage"
)

type Handlers struct {
	ClientHandler              *handler_client.ClientHandler
	ModelHandler               *handler_model.ModelHandler
	ServiceProviderHandler     *handler_service_provider.ServiceProviderHandler
	ClientModelRelationHandler *handler_client_model_relation.ClientModelRelationHandler
	PromptHandler              *handler_prompt.PromptHandler
	SessionHandler             *handler_session.SessionHandler
	ClientTokenHandler         *handler_client_token.ClientTokenHandler
	I18nHandler                *handler_i18n.I18nHandler
	RichClientHandler          *handler_rich_client.ClientHandler
	AuditHandler               *handler_audit.AuditHandler
	TokenUsageHandler          *handler_token_usage.TokenUsageHandler
	TemplateHandler            *handler_template.TemplateHandler
	PolicyGroupHandler         *handler_policy_group.Handler
	CacheHandler               *handler_cache.CacheHandler
}
