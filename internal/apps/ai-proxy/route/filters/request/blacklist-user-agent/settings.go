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

	settingpb "github.com/erda-project/erda-proto-go/apps/aiproxy/setting/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
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
	if cache, ok := ctxhelper.GetCacheManager(ctx); ok && cache != nil {
		if manager, ok := cache.(cachetypes.Manager); ok && manager != nil {
			if _, settingsV, err := manager.ListAll(ctx, cachetypes.ItemTypeSetting); err == nil {
				if list, ok := settingsV.([]*settingpb.Setting); ok {
					return buildSettingsFromList(list)
				}
			}
		}
	}

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

	list := make([]*settingpb.Setting, 0, len(items))
	for _, key := range []string{
		settingKeyClientTokenBlacklist,
		settingKeyClientBlacklist,
		settingKeyGeneralHeaders,
		settingKeyGeneralPrompts,
	} {
		if item, ok := items[key]; ok {
			list = append(list, item.ToProtobuf())
		}
	}

	return buildSettingsFromList(list)
}

func buildSettingsFromList(items []*settingpb.Setting) Settings {
	settings := Settings{}
	for _, item := range items {
		if item == nil || item.Namespace != blacklistUserAgentSettingNamespace {
			continue
		}
		switch item.Key {
		case settingKeyClientTokenBlacklist:
			settings.ClientTokenBlacklist = normalizeBlacklist(splitBlacklist(item.Value))
		case settingKeyClientBlacklist:
			settings.ClientBlacklist = normalizeBlacklist(splitBlacklist(item.Value))
		case settingKeyGeneralHeaders:
			settings.GeneralRules.Headers = normalizeGeneralRules(splitGeneralRules(item.Value))
		case settingKeyGeneralPrompts:
			settings.GeneralRules.Prompts = normalizeGeneralRules(splitGeneralRules(item.Value))
		}
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
