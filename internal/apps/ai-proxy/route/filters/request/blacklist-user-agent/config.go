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

type GeneralConfig struct {
	ItemTypesStr string   `json:"item_types_str" yaml:"item_types_str" file:"item_types_str"`
	ItemTypes    []string `json:"-" yaml:"-" file:"-"`
}

type Config struct {
	ClientToken ClientTokenConfig `json:"client_token" yaml:"client_token" file:"client_token"`
	Client      ClientConfig      `json:"client" yaml:"client" file:"client"`
	General     GeneralConfig     `json:"general" yaml:"general" file:"general"`
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
			Blacklist: normalizeBlacklist(resolveBlacklist(cfg.ClientToken.Blacklist, cfg.ClientToken.BlacklistStr)),
		},
		Client: ClientConfig{
			Blacklist: normalizeBlacklist(resolveBlacklist(cfg.Client.Blacklist, cfg.Client.BlacklistStr)),
		},
		General: GeneralConfig{
			ItemTypes: normalizeGeneralItemTypes(resolveGeneralItemTypes(cfg.General.ItemTypes, cfg.General.ItemTypesStr)),
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
		General: GeneralConfig{
			ItemTypes: append([]string(nil), currentConfig.General.ItemTypes...),
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

func resolveBlacklist(items []string, raw string) []string {
	if len(items) > 0 {
		return items
	}
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return strings.Split(raw, ",")
}

func normalizeGeneralItemTypes(items []string) []string {
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		cleaned := strings.TrimSpace(strings.Trim(strings.TrimSpace(item), ";"))
		if value := normalize(cleaned); value != "" {
			normalized = append(normalized, value)
		}
	}
	return normalized
}

func resolveGeneralItemTypes(items []string, raw string) []string {
	if len(items) > 0 {
		return items
	}
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return strings.Split(raw, ";;;")
}

func normalize(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}
