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
	"fmt"
	"sync"
)

type BlacklistItem interface {
	Name() string
}

type HeaderMatcher interface {
	MatchHeader(key, value string) bool
}

type PromptMatcher interface {
	MatchPrompt(prompt string) bool
}

type MessageGroupMatcher interface {
	MatchMessageGroupText(text string) bool
}

var (
	itemsMu         sync.RWMutex
	registeredItems = make(map[string]BlacklistItem)
)

func registerItem(item BlacklistItem) {
	itemsMu.Lock()
	defer itemsMu.Unlock()

	name := normalize(item.Name())
	if _, exists := registeredItems[name]; exists {
		panic(fmt.Errorf("blacklist user-agent item %s duplicated", name))
	}
	registeredItems[name] = item
}

func getItem(name string) (BlacklistItem, bool) {
	itemsMu.RLock()
	defer itemsMu.RUnlock()

	item, ok := registeredItems[normalize(name)]
	return item, ok
}

func resolveEnabledItems(blacklist []string) []BlacklistItem {
	items := make([]BlacklistItem, 0, len(blacklist))
	for _, itemName := range blacklist {
		item, ok := getItem(itemName)
		if !ok {
			continue
		}
		items = append(items, item)
	}
	return items
}

func resolveActiveItems(blacklist []string) []BlacklistItem {
	items := resolveEnabledItems(blacklist)
	if hasGeneralFallbackRules() {
		items = append(items, generalItem{})
	}
	return items
}

func hasGeneralFallbackRules() bool {
	cfg := getGeneralRules()
	return len(cfg.Headers) > 0 || len(cfg.Prompts) > 0
}
