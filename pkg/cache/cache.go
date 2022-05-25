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
	cachesC = make(chan *Cache, 1<<16)
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
	update  Update
	isSync  bool // call update synchronously when expired if true
}

// New returns the *Cache.
// the function update, if found new *item, returns true, and stores it;
// else returns false, and delete the key from cache.
func New(name string, expired time.Duration, update Update, options ...Option) *Cache {
	c := &Cache{
		C:       make(chan interface{}, 1<<16),
		name:    name,
		expired: expired,
		update:  update,
	}
	for _, option := range options {
		option(c)
	}
	return c
}

type Option func(c *Cache)

// WithSync set sync=true for Cache
func WithSync() Option {
	return func(c *Cache) {
		c.isSync = true
	}
}

// Name returns the Cache object's name
func (c *Cache) Name() string {
	return c.name
}

// LoadWithUpdate loads the cached item.
// If the item is not cached, it returns the newest and cache the new item.
// If the time is cached, it returns the item.
// if the cached item is expired, it tries to cache the newest.
func (c *Cache) LoadWithUpdate(key interface{}) (interface{}, bool) {
	value, ok := c.Map.Load(key)
	if !ok {
		obj, ok := c.update(key)
		if !ok {
			return nil, false
		}
		c.Store(key, obj)
		return obj, true
	}
	item := value.(*item)
	if item.expired.Before(time.Now()) {
		if c.isSync {
			obj, ok := c.update(key)
			if !ok {
				return nil, false
			}
			c.Store(key, obj)
			return obj, true
		}
		select {
		case c.C <- key:
			cachesC <- c
		default:
			logrus.WithField("func", "*Cache.LoadWithUpdate").
				WithField("name", c.Name()).Warnln("channel is blocked, update cache is skipped")
		}
	}
	return item.Object, true
}

// Store caches the key and value, and updates its expired time.
func (c *Cache) Store(key interface{}, value interface{}) {
	c.Map.Store(key, &item{Object: value, expired: time.Now().Add(c.expired)})
}

// item contains the item be cached
type item struct {
	Object interface{}

	expired time.Time
}

// Update retrieves the caching item
type Update func(interface{}) (interface{}, bool)
