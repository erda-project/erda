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

package cachetypes

import (
	"context"
	"sync"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

// IQueryFromDB defines the interface for database queries
type IQueryFromDB interface {
	QueryFromDB(ctx context.Context) (uint64, any, error)
	GetByIDFromData(ctx context.Context, data any, id string) (any, error)
}

// baseCacheItem provides the base implementation for CacheItem interface
type BaseCacheItem struct {
	data         any
	lastOK       bool
	mu           sync.RWMutex
	DBClient     dao.DAO
	itemType     ItemType       // identify the specific cache item type
	config       *config.Config // cache configuration
	IQueryFromDB                // embedded interface for database operations
}

func NewBaseCacheItem(itemType ItemType, dao dao.DAO, config *config.Config, queryFromDB IQueryFromDB) *BaseCacheItem {
	return &BaseCacheItem{
		DBClient:     dao,
		itemType:     itemType,
		config:       config,
		IQueryFromDB: queryFromDB,
	}
}

func (c *BaseCacheItem) getLogger(ctx context.Context) logs.Logger {
	l, ok := ctxhelper.GetLogger(ctx)
	if ok {
		return l
	}
	return logrusx.New().Sub("cache")
}

// ListAll implements the common logic for all cache items
func (c *BaseCacheItem) ListAll(ctx context.Context) (uint64, any, error) {
	// if cache is disabled, always query from database
	if !c.config.Enabled {
		c.getLogger(ctx).Warn("cache is disabled, fallback to database query")
		return c.QueryFromDB(ctx)
	}

	c.mu.RLock()
	if c.lastOK && c.data != nil {
		c.getLogger(ctx).Debugf("cache hit: %s", GetItemTypeName(c.itemType))
		data := c.data
		c.mu.RUnlock()
		return 0, data, nil
	}
	c.mu.RUnlock()

	// fallback to database query
	c.getLogger(ctx).Warnf("cache miss: %s, fallback to database query", GetItemTypeName(c.itemType))
	return c.QueryFromDB(ctx)
}

// GetByID implements the common logic for all cache items
func (c *BaseCacheItem) GetByID(ctx context.Context, id string) (any, error) {
	_, data, err := c.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return c.GetByIDFromData(ctx, data, id)
}

// Refresh implements the common logic for all cache items
func (c *BaseCacheItem) Refresh() (uint64, error) {
	total, data, err := c.QueryFromDB(context.Background())
	if err != nil {
		c.mu.Lock()
		c.lastOK = false
		c.mu.Unlock()
		return 0, err
	}

	c.mu.Lock()
	c.data = data
	c.lastOK = true
	c.mu.Unlock()
	return total, nil
}
