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

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

const (
	blacklistUserAgentSettingNamespace = "blacklist_user_agent"
	settingKeyClientTokenBlacklist     = "client_token.blacklist"
	settingKeyClientBlacklist          = "client.blacklist"
	settingKeyGeneralHeaders           = "general.headers"
	settingKeyGeneralPrompts           = "general.prompts"
)

type Settings struct {
	ClientTokenBlacklist []string
	ClientBlacklist      []string
	GeneralRules         GeneralRules
}

func resolveSettings(ctx context.Context) Settings {
	dbClient, ok := ctxhelper.GetDBClient(ctx)
	if !ok || dbClient == nil {
		return Settings{}
	}

	items, err := dbClient.SettingClient().GetByNamespaceKeys(
		ctx,
		blacklistUserAgentSettingNamespace,
		settingKeyClientTokenBlacklist,
		settingKeyClientBlacklist,
		settingKeyGeneralHeaders,
		settingKeyGeneralPrompts,
	)
	if err != nil {
		return Settings{}
	}

	settings := Settings{}
	if item, ok := items[settingKeyClientTokenBlacklist]; ok {
		settings.ClientTokenBlacklist = normalizeBlacklist(splitBlacklist(item.Value))
	}
	if item, ok := items[settingKeyClientBlacklist]; ok {
		settings.ClientBlacklist = normalizeBlacklist(splitBlacklist(item.Value))
	}
	if item, ok := items[settingKeyGeneralHeaders]; ok {
		settings.GeneralRules.Headers = normalizeGeneralRules(splitGeneralRules(item.Value))
	}
	if item, ok := items[settingKeyGeneralPrompts]; ok {
		settings.GeneralRules.Prompts = normalizeGeneralRules(splitGeneralRules(item.Value))
	}

	return settings
}

func getBlacklistByCredential(ctx context.Context, settings Settings) []string {
	switch {
	case auth.IsClientToken(ctx):
		return settings.ClientTokenBlacklist
	case auth.IsClient(ctx):
		return settings.ClientBlacklist
	default:
		return nil
	}
}
