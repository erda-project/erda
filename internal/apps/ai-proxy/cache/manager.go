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

package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

// getItemTypeName returns human-readable name for ItemType
func getItemTypeName(itemType cachetypes.ItemType) string {
	switch itemType {
	case cachetypes.ItemTypeModel:
		return "model"
	case cachetypes.ItemTypeProvider:
		return "provider"
	default:
		return fmt.Sprintf("unknown(%d)", itemType)
	}
}

// cacheManager implements the Manager interface using CacheItem
type cacheManager struct {
	items  map[cachetypes.ItemType]cachetypes.CacheItem
	config *config
	logger logs.Logger
}

// NewCacheManager creates a new cache manager instance
func NewCacheManager(dao dao.DAO, logger logs.Logger, isMcpProxy bool) cachetypes.Manager {
	config := loadConfig()
	manager := &cacheManager{
		items:  map[cachetypes.ItemType]cachetypes.CacheItem{},
		config: config,
		logger: logger,
	}
	if !isMcpProxy {
		manager.items[cachetypes.ItemTypeModel] = newModelCacheItem(dao, config)
		manager.items[cachetypes.ItemTypeProvider] = newProviderCacheItem(dao, config)
	}

	// start background refresh goroutine only if cache is enabled
	if config.Enabled {
		go manager.startRefreshLoop()
	}

	return manager
}

// ListAll returns all cached items of the specified type
func (m *cacheManager) ListAll(ctx context.Context, itemType cachetypes.ItemType) (any, error) {
	item := m.items[itemType]
	if item == nil {
		return nil, fmt.Errorf("unsupported item type: %d", itemType)
	}

	return item.ListAll(ctx)
}

// GetByID returns an item by ID for the specified type
func (m *cacheManager) GetByID(ctx context.Context, itemType cachetypes.ItemType, id string) (any, error) {
	item := m.items[itemType]
	if item == nil {
		return nil, fmt.Errorf("unsupported item type: %d", itemType)
	}

	return item.GetByID(ctx, id)
}

// startRefreshLoop starts the background refresh loop
func (m *cacheManager) startRefreshLoop() {
	// start independent refresh goroutine for each item type
	for itemType, item := range m.items {
		go func(t cachetypes.ItemType, i cachetypes.CacheItem) {
			for {
				if err := i.Refresh(); err != nil {
					m.logger.Errorf("cache refresh error for %s: %v (interval: %s)", getItemTypeName(t), err, m.config.RefreshInterval)
				} else {
					m.logger.Infof("cache refresh success for %s (interval: %s)", getItemTypeName(t), m.config.RefreshInterval)
				}
				time.Sleep(m.config.RefreshInterval)
			}
		}(itemType, item)
	}
}
