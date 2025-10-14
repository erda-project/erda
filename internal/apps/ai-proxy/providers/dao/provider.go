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

package dao

import (
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/audit"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_mcp_relation"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_model_relation"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/i18n"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_server"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/model"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/model_provider"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/prompt"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/session"
	usage_token "github.com/erda-project/erda/internal/apps/ai-proxy/models/usage/token"
)

var (
	_ DAO = (*provider)(nil)
)

var (
	name = "erda.apps.ai-proxy.dao"
	spec = servicehub.Spec{
		Services:    []string{"erda.apps.ai-proxy.dao"},
		Summary:     "erda.apps.ai-proxy.dao",
		Description: "erda.apps.ai-proxy.dao",
		ConfigFunc: func() any {
			return new(struct{})
		},
		Types: pb.Types(),
		Creator: func() servicehub.Provider {
			return new(provider)
		},
	}
)

func init() {
	servicehub.Register(name, &spec)
}

type DAO interface {
	Q() *gorm.DB
	Tx() *gorm.DB

	ModelProviderClient() *model_provider.DBClient
	ClientClient() *client.DBClient
	ModelClient() *model.DBClient
	ClientModelRelationClient() *client_model_relation.DBClient
	PromptClient() *prompt.DBClient
	SessionClient() *session.DBClient
	ClientTokenClient() *client_token.DBClient
	AuditClient() *audit.DBClient
	MCPServerClient() *mcp_server.DBClient
	I18nClient() *i18n.DBClient
	ClientMCPRelationClient() *client_mcp_relation.DBClient
	TokenUsageClient() *usage_token.DBClient
}

type provider struct {
	DB *gorm.DB `autowired:"mysql-gorm.v2-client"`
}

func (p *provider) Provide(ctx servicehub.DependencyContext, options ...any) any {
	return p
}

func (p *provider) Q() *gorm.DB {
	return p.DB
}

func (p *provider) Tx() *gorm.DB {
	return p.DB.Session(&gorm.Session{})
}

func (p *provider) ModelProviderClient() *model_provider.DBClient {
	return &model_provider.DBClient{DB: p.DB}
}

func (p *provider) ClientClient() *client.DBClient {
	return &client.DBClient{DB: p.DB}
}

func (p *provider) ModelClient() *model.DBClient {
	return &model.DBClient{DB: p.DB}
}

func (p *provider) ClientModelRelationClient() *client_model_relation.DBClient {
	return &client_model_relation.DBClient{DB: p.DB}
}

func (p *provider) PromptClient() *prompt.DBClient {
	return &prompt.DBClient{DB: p.DB}
}

func (p *provider) SessionClient() *session.DBClient {
	return &session.DBClient{DB: p.DB}
}

func (p *provider) ClientTokenClient() *client_token.DBClient {
	return &client_token.DBClient{DB: p.DB}
}

func (p *provider) AuditClient() *audit.DBClient {
	return &audit.DBClient{DB: p.DB}
}

func (p *provider) MCPServerClient() *mcp_server.DBClient {
	return &mcp_server.DBClient{DB: p.DB}
}

func (p *provider) I18nClient() *i18n.DBClient {
	return &i18n.DBClient{DB: p.DB}
}

func (p *provider) ClientMCPRelationClient() *client_mcp_relation.DBClient {
	return &client_mcp_relation.DBClient{DB: p.DB}
}

func (p *provider) TokenUsageClient() *usage_token.DBClient {
	return &usage_token.DBClient{DB: p.DB}
}
