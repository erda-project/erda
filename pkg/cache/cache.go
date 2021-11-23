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

	"github.com/sirupsen/logrus"
)

var (
	cachesC = make(chan *Cache, 100)
)

func init() {
	go asyncUpdate()
}

func asyncUpdate() {
	for {
		select {
		case c := <-cachesC:
			k := <-c.C
			if v, ok := c.update(k); ok {
				c.Store(k, v)
			} else {
				c.Delete(k)
			}
		}
	}
}

// Cache defines the methods LoadWithUpdate and Store,
// the user needs to define the function `update`.
type Cache struct {
	sync.Map

	C       chan interface{}
	name    string
	expired time.Duration
	update  func(interface{}) (*Item, bool)
}

// New returns the *Cache.
// the function update, if found new *Item, returns true, and stores it;
// else returns false, and delete the key from cache.
func New(name string, expired time.Duration, update func(interface{}) (*Item, bool)) *Cache {
	return &Cache{
		C:       make(chan interface{}, 1000),
		name:    name,
		expired: expired,
		update:  update,
	}
}

// Name returns the Cache object's name
func (c *Cache) Name() string {
	return c.name
}

// LoadWithUpdate loads the cached item.
// if the item is not cached, it returns the newest and cache the new item.
// if the item is expired, it returns the cached item and try to cache the newest.
// if the item is not expired, it returns it and do nothing.
func (c *Cache) LoadWithUpdate(key interface{}) (*Item, bool) {
	value, ok := c.Map.Load(key)
	if !ok {
		return c.update(key)
	}
	item := value.(*Item)
	if item.IsExpired() {
		select {
		case c.C <- key:
			cachesC <- c
		default:
			logrus.WithField("func", "*Cache.LoadWithUpdate").
				WithField("name", c.Name()).Warnln("channel is blocked, update cache is skipped")
		}
	}
	return item, true
}

// Store caches the key and value, and updates its expired time.
func (c *Cache) Store(key interface{}, value *Item) {
	c.Map.Store(key, value)
	value.updateExpired(c.expired)
}

// Item contains the item be cached
type Item struct {
	Object interface{}

	expired time.Time
}

// IsExpired returns whether the item is expired
func (i *Item) IsExpired() bool {
	return i.expired.Before(time.Now())
}

func (i *Item) updateExpired(duration time.Duration) {
	i.expired = time.Now().Add(duration)
}
