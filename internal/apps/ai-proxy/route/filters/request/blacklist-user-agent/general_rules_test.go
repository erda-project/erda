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

package blacklist_user_agent

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/audit"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_mcp_relation"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_model_relation"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/i18n"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_server"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_server_config_instance"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_server_template"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/model"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/prompt"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/service_provider"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/session"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/setting"
	usage_token "github.com/erda-project/erda/internal/apps/ai-proxy/models/usage/token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

func TestResolveSettings_PrefersConfiguredGeneralRules(t *testing.T) {
	ctx := newContextWithSettingDAO(t)
	dbClient := ctxhelper.MustGetDBClient(ctx).SettingClient()
	require.NoError(t, dbClient.CreateOrUpdate(ctx, &setting.Setting{
		Namespace: blacklistUserAgentSettingNamespace,
		Key:       settingKeyGeneralHeaders,
		Value:     "claude code",
	}))
	require.NoError(t, dbClient.CreateOrUpdate(ctx, &setting.Setting{
		Namespace: blacklistUserAgentSettingNamespace,
		Key:       settingKeyGeneralPrompts,
		Value:     "You are Claude Code",
	}))

	got := resolveSettings(ctx)
	require.Equal(t, []string{"claude code"}, got.GeneralRules.Headers)
	require.Equal(t, []string{"you are claude code"}, got.GeneralRules.Prompts)
}

func TestResolveSettings_ReturnsEmptyWhenSettingsMissing(t *testing.T) {
	ctx := newContextWithSettingDAO(t)

	got := resolveSettings(ctx)
	require.Empty(t, got.ClientTokenBlacklist)
	require.Empty(t, got.ClientBlacklist)
	require.Empty(t, got.GeneralRules.Headers)
	require.Empty(t, got.GeneralRules.Prompts)
}

func TestResolveSettings_UsesOnlyConfiguredSettingKeys(t *testing.T) {
	ctx := newContextWithSettingDAO(t)
	dbClient := ctxhelper.MustGetDBClient(ctx).SettingClient()
	require.NoError(t, dbClient.CreateOrUpdate(ctx, &setting.Setting{
		Namespace: blacklistUserAgentSettingNamespace,
		Key:       settingKeyGeneralHeaders,
		Value:     "claude code",
	}))

	got := resolveSettings(ctx)
	require.Equal(t, []string{"claude code"}, got.GeneralRules.Headers)
	require.Empty(t, got.GeneralRules.Prompts)
}

func TestResolveSettings_UsesSettingValues(t *testing.T) {
	ctx := newContextWithSettingDAO(t)
	dbClient := ctxhelper.MustGetDBClient(ctx).SettingClient()
	require.NoError(t, dbClient.CreateOrUpdate(ctx, &setting.Setting{
		Namespace: blacklistUserAgentSettingNamespace,
		Key:       settingKeyClientTokenBlacklist,
		Value:     " openclaw, Coding-Agent ,, ",
	}))
	require.NoError(t, dbClient.CreateOrUpdate(ctx, &setting.Setting{
		Namespace: blacklistUserAgentSettingNamespace,
		Key:       settingKeyClientBlacklist,
		Value:     "coding-agent, openclaw",
	}))
	require.NoError(t, dbClient.CreateOrUpdate(ctx, &setting.Setting{
		Namespace: blacklistUserAgentSettingNamespace,
		Key:       settingKeyGeneralHeaders,
		Value:     "claude code ;;; opencode",
	}))
	require.NoError(t, dbClient.CreateOrUpdate(ctx, &setting.Setting{
		Namespace: blacklistUserAgentSettingNamespace,
		Key:       settingKeyGeneralPrompts,
		Value:     "You are Claude Code ;;; You are OpenCode",
	}))

	got := resolveSettings(ctx)
	require.Equal(t, []string{"openclaw", "coding-agent"}, got.ClientTokenBlacklist)
	require.Equal(t, []string{"coding-agent", "openclaw"}, got.ClientBlacklist)
	require.Equal(t, []string{"claude code", "opencode"}, got.GeneralRules.Headers)
	require.Equal(t, []string{"you are claude code", "you are opencode"}, got.GeneralRules.Prompts)
}

func newContextWithSettingDAO(t *testing.T) context.Context {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, prepareSQLiteSettingTableForBlacklist(db))

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutDBClient(ctx, testDAO{db: db})
	return ctx
}

func prepareSQLiteSettingTableForBlacklist(db *gorm.DB) error {
	return db.Exec(`
CREATE TABLE ai_proxy_setting (
	id CHAR(36) PRIMARY KEY,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	deleted_at DATETIME NULL,
	namespace VARCHAR(191) NOT NULL,
	key VARCHAR(191) NOT NULL,
	value TEXT NOT NULL DEFAULT '',
	UNIQUE(namespace, key)
);`).Error
}

type testDAO struct {
	db *gorm.DB
}

func (d testDAO) Q() *gorm.DB { return d.db }
func (d testDAO) Tx() *gorm.DB { return d.db }
func (d testDAO) ServiceProviderClient() *service_provider.DBClient { return nil }
func (d testDAO) ClientClient() *client.DBClient { return nil }
func (d testDAO) ModelClient() *model.DBClient { return nil }
func (d testDAO) ClientModelRelationClient() *client_model_relation.DBClient { return nil }
func (d testDAO) PromptClient() *prompt.DBClient { return nil }
func (d testDAO) SessionClient() *session.DBClient { return nil }
func (d testDAO) ClientTokenClient() *client_token.DBClient { return nil }
func (d testDAO) AuditClient() *audit.DBClient { return nil }
func (d testDAO) MCPServerClient() *mcp_server.DBClient { return nil }
func (d testDAO) I18nClient() *i18n.DBClient { return nil }
func (d testDAO) ClientMCPRelationClient() *client_mcp_relation.DBClient { return nil }
func (d testDAO) TokenUsageClient() *usage_token.DBClient { return nil }
func (d testDAO) MCPServerTemplateClient() *mcp_server_template.DBClient { return nil }
func (d testDAO) MCPServerConfigInstanceClient() *mcp_server_config_instance.DBClient { return nil }
func (d testDAO) PolicyGroupClient() *policy_group.DBClient { return nil }
func (d testDAO) SettingClient() *setting.DBClient { return &setting.DBClient{DB: d.db} }

var _ dao.DAO = testDAO{}
