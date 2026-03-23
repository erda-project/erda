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
	"sync"
)

type BlacklistItem interface {
	Name() string
	Match(ctx context.Context) (bool, string)
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
