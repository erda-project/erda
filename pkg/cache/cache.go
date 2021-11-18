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
	"sync"
	"time"
)

func RunCache(caches ...*Cache) {
	go func() {
		for {
			for _, c := range caches {
				select {
				case k := <-c.C:
					if v, ok := c.update(k); ok {
						c.Store(k, v)
					}
				default:
				}
			}
		}
	}()
}

type Cache struct {
	sync.Map
	expired time.Duration
	C       chan interface{}
	update  func(interface{}) (*CacheItem, bool)
}

func New(expired time.Duration, update func(interface{}) (*CacheItem, bool)) *Cache {
	return &Cache{
		expired: expired,
		C:       make(chan interface{}, 1000),
		update:  update,
	}
}

func (c *Cache) LoadWithUpdate(key interface{}) (*CacheItem, bool) {
	value, ok := c.Map.Load(key)
	if !ok {
		return c.update(key)
	}
	item := value.(*CacheItem)
	if item.IsExpired() {
		return c.update(key)
	}
	return item, true
}

func (c *Cache) Store(key interface{}, value *CacheItem) {
	c.Map.Store(key, value)
	value.updateExpiredTime(c.expired)
}

type CacheItem struct {
	expiredTime time.Time

	Object interface{}
}

func (i *CacheItem) IsExpired() bool {
	return i.expiredTime.Before(time.Now())
}

func (i *CacheItem) updateExpiredTime(duration time.Duration) {
	i.expiredTime = time.Now().Add(duration)
}
