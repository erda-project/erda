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
	BlacklistStr string   `json:"blacklist_str" yaml:"blacklist_str" file:"blacklist_str"`
	Blacklist    []string `json:"-" yaml:"-" file:"-"`
}

type ClientConfig struct {
	BlacklistStr string   `json:"blacklist_str" yaml:"blacklist_str" file:"blacklist_str"`
	Blacklist    []string `json:"-" yaml:"-" file:"-"`
}

type Config struct {
	ClientToken ClientTokenConfig `json:"client_token" yaml:"client_token" file:"client_token"`
	Client      ClientConfig      `json:"client" yaml:"client" file:"client"`
}

type GeneralRules struct {
	Headers []string
	Prompts []string
}

var (
	configMu            sync.RWMutex
	currentConfig       Config
	currentGeneralRules GeneralRules
)

func SetConfig(cfg Config) {
	configMu.Lock()
	defer configMu.Unlock()

	currentConfig = Config{
		ClientToken: ClientTokenConfig{
			Blacklist: normalizeBlacklist(resolveBlacklist(cfg.ClientToken.Blacklist, cfg.ClientToken.BlacklistStr)),
		},
		Client: ClientConfig{
			Blacklist: normalizeBlacklist(resolveBlacklist(cfg.Client.Blacklist, cfg.Client.BlacklistStr)),
		},
	}
}

func SetGeneralRules(headersRaw, promptsRaw string) {
	configMu.Lock()
	defer configMu.Unlock()

	currentGeneralRules = GeneralRules{
		Headers: normalizeGeneralRules(splitGeneralRules(headersRaw)),
		Prompts: normalizeGeneralRules(splitGeneralRules(promptsRaw)),
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

func getGeneralRules() GeneralRules {
	configMu.RLock()
	defer configMu.RUnlock()

	return GeneralRules{
		Headers: append([]string(nil), currentGeneralRules.Headers...),
		Prompts: append([]string(nil), currentGeneralRules.Prompts...),
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

func resolveBlacklist(items []string, raw string) []string {
	if len(items) > 0 {
		return items
	}
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return strings.Split(raw, ",")
}

func normalizeGeneralRules(items []string) []string {
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		cleaned := strings.TrimSpace(strings.Trim(strings.TrimSpace(item), ";"))
		if value := normalize(cleaned); value != "" {
			normalized = append(normalized, value)
		}
	}
	return normalized
}

func splitGeneralRules(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return strings.Split(raw, ";;;")
}

func normalize(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}
