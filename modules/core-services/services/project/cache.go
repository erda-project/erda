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

package project

import (
	"sync"
	"time"
)

type Cache struct {
	sync.Map
	expiredDuration time.Duration
	c               chan struct{}
}

func NewCache(expiredDuration time.Duration) *Cache {
	return &Cache{expiredDuration: expiredDuration, c: make(chan struct{}, 10)}
}

func (c *Cache) Store(key interface{}, value interface{}) {
	c.Map.Store(key, value)
	if v, ok := value.(CacheItme); ok {
		v.UpdateExpiredTime(c.expiredDuration)
	}
}

func (c *Cache) Lock() {
	c.c <- struct{}{}
}

func (c *Cache) Release() {
	<-c.c
}

type CacheItme interface {
	IsExpired() bool
	UpdateExpiredTime(duration time.Duration)
}

type quotaCache struct {
	ProjectID          uint64
	ProjectName        string
	ProjectDisplayName string
	ProjectDesc        string
	CPUQuota           uint64
	MemQuota           uint64

	expiredTime time.Time
}

func (i *quotaCache) IsExpired() bool {
	return i.expiredTime.Before(time.Now())
}

func (i *quotaCache) UpdateExpiredTime(duration time.Duration) {
	i.expiredTime = time.Now().Add(duration)
}

type memberCache struct {
	ProjectID uint64
	UserID    uint
	Name      string
	Nick      string

	expiredTime time.Time
}

func (i *memberCache) IsExpired() bool {
	return i.expiredTime.Before(time.Now())
}

func (i *memberCache) UpdateExpiredTime(duration time.Duration) {
	i.expiredTime = time.Now().Add(duration)
}
