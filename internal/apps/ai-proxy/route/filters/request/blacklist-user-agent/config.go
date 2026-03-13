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
	"strings"
	"sync"
)

type ClientTokenConfig struct {
	Blacklist []string `json:"blacklist" yaml:"blacklist" file:"blacklist" env:"AI_PROXY_BLACKLIST_USER_AGENT_FOR_CLIENT_TOKEN"`
}

type ClientConfig struct {
	Blacklist []string `json:"blacklist" yaml:"blacklist" file:"blacklist" env:"AI_PROXY_BLACKLIST_USER_AGENT_FOR_CLIENT"`
}

type Config struct {
	ClientToken ClientTokenConfig `json:"client_token" yaml:"client_token" file:"client_token"`
	Client      ClientConfig      `json:"client" yaml:"client" file:"client"`
}

var (
	configMu      sync.RWMutex
	currentConfig Config
)

func SetConfig(cfg Config) {
	configMu.Lock()
	defer configMu.Unlock()

	currentConfig = Config{
		ClientToken: ClientTokenConfig{
			Blacklist: normalizeBlacklist(cfg.ClientToken.Blacklist),
		},
		Client: ClientConfig{
			Blacklist: normalizeBlacklist(cfg.Client.Blacklist),
		},
	}
}

func getConfig() Config {
	configMu.RLock()
	defer configMu.RUnlock()

	return Config{
		ClientToken: ClientTokenConfig{
			Blacklist: append([]string(nil), currentConfig.ClientToken.Blacklist...),
		},
		Client: ClientConfig{
			Blacklist: append([]string(nil), currentConfig.Client.Blacklist...),
		},
	}
}

func normalizeBlacklist(items []string) []string {
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		if value := normalize(item); value != "" {
			normalized = append(normalized, value)
		}
	}
	return normalized
}

func normalize(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}
