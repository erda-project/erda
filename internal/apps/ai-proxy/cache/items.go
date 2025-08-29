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
	"sync"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/logs/logrusx"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

// QueryFromDB defines the interface for database queries
type QueryFromDB interface {
	queryFromDB(ctx context.Context) (any, error)
	getByIDFromData(ctx context.Context, data any, id string) (any, error)
}

// baseCacheItem provides the base implementation for CacheItem interface
type baseCacheItem struct {
	data        any
	lastOK      bool
	mu          sync.RWMutex
	dao         dao.DAO
	itemType    string  // identify the specific cache item type
	config      *config // cache configuration
	QueryFromDB         // embedded interface for database operations
}

func (c *baseCacheItem) getLogger(ctx context.Context) logs.Logger {
	l, ok := ctxhelper.GetLogger(ctx)
	if ok {
		return l
	}
	return logrusx.New().Sub("cache")
}

// ListAll implements the common logic for all cache items
func (c *baseCacheItem) ListAll(ctx context.Context) (any, error) {
	// if cache is disabled, always query from database
	if !c.config.Enabled {
		c.getLogger(ctx).Warn("cache is disabled, fallback to database query")
		return c.queryFromDB(ctx)
	}

	c.mu.RLock()
	if c.lastOK && c.data != nil {
		c.getLogger(ctx).Infof("cache hit: %s", c.itemType)
		data := c.data
		c.mu.RUnlock()
		return data, nil
	}
	c.mu.RUnlock()

	// fallback to database query
	c.getLogger(ctx).Warnf("cache miss: %s, fallback to database query", c.itemType)
	return c.queryFromDB(ctx)
}

// GetByID implements the common logic for all cache items
func (c *baseCacheItem) GetByID(ctx context.Context, id string) (any, error) {
	data, err := c.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return c.getByIDFromData(ctx, data, id)
}

// Refresh implements the common logic for all cache items
func (c *baseCacheItem) Refresh() error {
	data, err := c.queryFromDB(context.Background())
	if err != nil {
		c.mu.Lock()
		c.lastOK = false
		c.mu.Unlock()
		return err
	}

	c.mu.Lock()
	c.data = data
	c.lastOK = true
	c.mu.Unlock()
	return nil
}

// modelCacheItem implements CacheItem for models
type modelCacheItem struct {
	*baseCacheItem
}

func newModelCacheItem(dao dao.DAO, config *config) cachetypes.CacheItem {
	item := &modelCacheItem{}
	item.baseCacheItem = &baseCacheItem{
		dao:         dao,
		itemType:    "model",
		config:      config,
		QueryFromDB: item, // self-reference for interface implementation
	}
	return item
}

func (c *modelCacheItem) queryFromDB(ctx context.Context) (any, error) {
	resp, err := c.dao.ModelClient().Paging(ctx, &modelpb.ModelPagingRequest{
		PageNum:  1,
		PageSize: 999,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query models: %w", err)
	}
	return resp.List, nil
}

func (c *modelCacheItem) getByIDFromData(ctx context.Context, data any, id string) (any, error) {
	models, ok := data.([]*modelpb.Model)
	if !ok {
		return nil, fmt.Errorf("invalid model data type")
	}

	for _, model := range models {
		if model.Id == id {
			return model, nil
		}
	}

	return nil, fmt.Errorf("model with ID %s not found", id)
}

// providerCacheItem implements CacheItem for providers
type providerCacheItem struct {
	*baseCacheItem
}

func newProviderCacheItem(dao dao.DAO, config *config) cachetypes.CacheItem {
	item := &providerCacheItem{}
	item.baseCacheItem = &baseCacheItem{
		dao:         dao,
		itemType:    "provider",
		config:      config,
		QueryFromDB: item, // self-reference for interface implementation
	}
	return item
}

func (c *providerCacheItem) queryFromDB(ctx context.Context) (any, error) {
	resp, err := c.dao.ModelProviderClient().Paging(ctx, &providerpb.ModelProviderPagingRequest{
		PageNum:  1,
		PageSize: 999,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query providers: %w", err)
	}
	return resp.List, nil
}

func (c *providerCacheItem) getByIDFromData(ctx context.Context, data any, id string) (any, error) {
	providers, ok := data.([]*providerpb.ModelProvider)
	if !ok {
		return nil, fmt.Errorf("invalid provider data type")
	}

	for _, provider := range providers {
		if provider.Id == id {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("provider with ID %s not found", id)
}
