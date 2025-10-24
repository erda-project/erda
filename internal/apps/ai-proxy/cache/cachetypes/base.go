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
	"fmt"
	"reflect"
	"sync"

	"github.com/mohae/deepcopy"
	"google.golang.org/protobuf/proto"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

// IQueryFromDB defines the interface for database queries
type IQueryFromDB interface {
	QueryFromDB(ctx context.Context) (uint64, any, error)
	GetIDValue(item any) (string, error)
}

// baseCacheItem provides the base implementation for CacheItem interface
type BaseCacheItem struct {
	data         any
	lastOK       bool
	index        map[string]any
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
	return c.listAll(ctx, true)
}

func (c *BaseCacheItem) listAll(ctx context.Context, needClone bool) (uint64, any, error) {
	// if cache is disabled, always query from database
	if !c.config.Enabled {
		c.getLogger(ctx).Warn("cache is disabled, fallback to database query")
		return c.QueryFromDB(ctx)
	}

	c.mu.RLock()
	cached := c.lastOK && c.data != nil
	c.mu.RUnlock()

	if cached {
		c.mu.Lock()
		if c.lastOK && c.data != nil {
			if c.index == nil {
				index, err := c.buildIndexFromListData(c.data)
				if err != nil {
					c.mu.Unlock()
					return 0, nil, err
				}
				c.index = index
			}
			data := c.data
			if needClone {
				data = smartClone(c.data)
			}
			c.mu.Unlock()
			c.getLogger(ctx).Debugf("cache hit: %s", GetItemTypeName(c.itemType))
			return 0, data, nil
		}
		c.mu.Unlock()
	}

	// fallback to database query
	c.getLogger(ctx).Warnf("cache miss: %s, fallback to database query", GetItemTypeName(c.itemType))
	total, newData, err := c.QueryFromDB(ctx)
	if err != nil {
		return 0, nil, err
	}

	index, err := c.buildIndexFromListData(newData)
	if err != nil {
		return 0, nil, err
	}

	resultData := newData
	if needClone {
		resultData = smartClone(newData)
	}

	c.mu.Lock()
	c.data = newData
	c.index = index
	c.lastOK = true
	c.mu.Unlock()
	return total, resultData, nil
}

func (c *BaseCacheItem) buildIndexFromListData(data any) (map[string]any, error) {
	index := make(map[string]any)
	if data == nil {
		return index, nil
	}

	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return index, nil
		}
		value = value.Elem()
	}

	if value.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected slice data for %s cache item", GetItemTypeName(c.itemType))
	}

	length := value.Len()
	for i := 0; i < length; i++ {
		elem := value.Index(i).Interface()
		id, err := c.GetIDValue(elem)
		if err != nil {
			return nil, err
		}
		if id == "" {
			return nil, fmt.Errorf("empty id for %s cache item", GetItemTypeName(c.itemType))
		}
		index[id] = elem
	}
	return index, nil
}

// smartClone:
// - if type is slice, clone each element
// - if type is proto.Message, use proto.Clone
// - otherwise, use deepcopy.Copy
func smartClone(data any) any {
	if data == nil {
		return nil
	}
	if msg, ok := data.(proto.Message); ok {
		return proto.Clone(msg)
	}

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Slice {
		if val.IsNil() {
			return nil
		}

		length := val.Len()
		cloned := reflect.MakeSlice(val.Type(), length, length)
		for i := 0; i < length; i++ {
			elem := val.Index(i).Interface()
			clonedElem := smartClone(elem)
			if clonedElem == nil {
				cloned.Index(i).Set(reflect.Zero(val.Type().Elem()))
				continue
			}
			elemValue := reflect.ValueOf(clonedElem)
			if !elemValue.Type().AssignableTo(val.Type().Elem()) {
				elemValue = elemValue.Convert(val.Type().Elem())
			}
			cloned.Index(i).Set(elemValue)
		}
		return cloned.Interface()
	}

	return deepcopy.Copy(data)
}

// GetByID implements the common logic for all cache items
func (c *BaseCacheItem) GetByID(ctx context.Context, id string) (any, error) {
	if !c.config.Enabled {
		_, data, err := c.listAll(ctx, false)
		if err != nil {
			return nil, err
		}
		return c.getByIDFromData(ctx, data, id)
	}

	if _, _, err := c.listAll(ctx, false); err != nil {
		return nil, err
	}

	c.mu.RLock()
	v, ok := c.index[id]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%s with ID %s not found", GetItemTypeName(c.itemType), id)
	}
	return smartClone(v), nil
}

func (c *BaseCacheItem) getByIDFromData(ctx context.Context, data any, id string) (any, error) {
	index, err := c.buildIndexFromListData(data)
	if err != nil {
		return nil, err
	}
	v, ok := index[id]
	if !ok {
		return nil, fmt.Errorf("%s with ID %s not found", GetItemTypeName(c.itemType), id)
	}
	return smartClone(v), nil
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

	index, err := c.buildIndexFromListData(data)
	if err != nil {
		c.mu.Lock()
		c.lastOK = false
		c.mu.Unlock()
		return 0, err
	}

	c.mu.Lock()
	c.data = data
	c.index = index
	c.lastOK = true
	c.mu.Unlock()
	return total, nil
}
