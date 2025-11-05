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
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cacheimpl/item_client"
	item_clienttoken "github.com/erda-project/erda/internal/apps/ai-proxy/cache/cacheimpl/item_client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cacheimpl/item_clientmodelrelation"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cacheimpl/item_model"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cacheimpl/item_provider"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cacheimpl/item_template"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template/templatetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

// cacheManager implements the Manager interface using CacheItem
type cacheManager struct {
	items  map[cachetypes.ItemType]cachetypes.CacheItem
	config *config.Config
	logger logs.Logger
}

// NewCacheManager creates a new cache manager instance
func NewCacheManager(dao dao.DAO, logger logs.Logger, templatesByType templatetypes.TemplatesByType, isMcpProxy bool) cachetypes.Manager {
	cfg := config.LoadConfig()
	manager := &cacheManager{
		items:  map[cachetypes.ItemType]cachetypes.CacheItem{},
		config: cfg,
		logger: logger,
	}
	manager.items[cachetypes.ItemTypeClient] = item_client.NewClientCacheItem(dao, cfg)
	if !isMcpProxy {
		manager.items[cachetypes.ItemTypeModel] = item_model.NewModelCacheItem(dao, cfg)
		manager.items[cachetypes.ItemTypeProvider] = item_provider.NewProviderCacheItem(dao, cfg)
		manager.items[cachetypes.ItemTypeClientModelRelation] = item_clientmodelrelation.NewClientModelRelationCacheItem(dao, cfg)
		manager.items[cachetypes.ItemTypeClientToken] = item_clienttoken.NewClientTokenCacheItem(dao, cfg)
		manager.items[cachetypes.ItemTypeTemplate] = item_template.NewTemplateCacheItem(dao, cfg, templatesByType)
	}

	// start background refresh goroutine only if cache is enabled
	if cfg.Enabled {
		go manager.startRefreshLoop()
	}

	return manager
}

func tryGetID(ctx context.Context) string {
	callID, ok := ctxhelper.GetGeneratedCallID(ctx)
	if ok {
		return fmt.Sprintf("call: %s", callID)
	}
	return fmt.Sprintf("uuid: %s", uuid.New())
}

// ListAll returns all cached items of the specified type
func (m *cacheManager) ListAll(ctx context.Context, itemType cachetypes.ItemType) (uint64, any, error) {
	cacheReqID := tryGetID(ctx)
	m.logger.Debugf("[%s] cache in: list all cache items for %s", cacheReqID, cachetypes.GetItemTypeName(itemType))
	defer m.logger.Debugf("[%s] cache out: list all cache items for %s", cacheReqID, cachetypes.GetItemTypeName(itemType))

	item := m.items[itemType]
	if item == nil {
		return 0, nil, fmt.Errorf("unsupported item type: %d", itemType)
	}

	return item.ListAll(ctx)
}

// GetByID returns an item by ID for the specified type
func (m *cacheManager) GetByID(ctx context.Context, itemType cachetypes.ItemType, id string) (any, error) {
	cacheReqID := tryGetID(ctx)
	m.logger.Debugf("[%s] cache in: get one cache item for %s, id: %s", cacheReqID, cachetypes.GetItemTypeName(itemType), id)
	defer m.logger.Debugf("[%s] cache out: get one cache item for %s, id: %s", cacheReqID, cachetypes.GetItemTypeName(itemType), id)

	item := m.items[itemType]
	if item == nil {
		return nil, fmt.Errorf("unsupported item type: %d", itemType)
	}

	return item.GetByID(ctx, id)
}

// TriggerRefresh triggers a refresh for the specified item
func (m *cacheManager) TriggerRefresh(ctx context.Context, itemTypes ...cachetypes.ItemType) {
	for _, t := range itemTypes {
		i := m.items[t]
		m.refreshOneItemType(t, i)
	}
}

// startRefreshLoop starts the background refresh loop
func (m *cacheManager) startRefreshLoop() {
	// start the independent refresh goroutine for each item type
	for itemType, item := range m.items {
		go func(t cachetypes.ItemType, i cachetypes.CacheItem) {
			for {
				m.refreshOneItemType(t, i)
				time.Sleep(m.config.RefreshInterval)
			}
		}(itemType, item)
	}
}

func (m *cacheManager) refreshOneItemType(t cachetypes.ItemType, i cachetypes.CacheItem) {
	timeBegin := time.Now()
	total, err := i.Refresh()
	timeCost := time.Since(timeBegin)
	if err != nil {
		m.logger.Errorf("cache refresh error for %s: %v (interval: %s) (cost: %s)", cachetypes.GetItemTypeName(t), err, m.config.RefreshInterval, timeCost)
	} else {
		m.logger.Infof("cache refresh success for %s (interval: %s) (cost: %s), total: %d", cachetypes.GetItemTypeName(t), m.config.RefreshInterval, timeCost, total)
	}
}
