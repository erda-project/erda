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

type Config struct {
	Blacklist []string `json:"blacklist" yaml:"blacklist" file:"blacklist" env:"AI_PROXY_BLACKLIST_USER_AGENT"`
}

var (
	configMu      sync.RWMutex
	currentConfig Config
)

func SetConfig(cfg Config) {
	configMu.Lock()
	defer configMu.Unlock()

	currentConfig = Config{
		Blacklist: normalizeBlacklist(cfg.Blacklist),
	}
}

func getConfig() Config {
	configMu.RLock()
	defer configMu.RUnlock()

	return Config{
		Blacklist: append([]string(nil), currentConfig.Blacklist...),
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
