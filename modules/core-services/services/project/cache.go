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
	C               chan uint64
}

func NewCache(expiredDuration time.Duration) *Cache {
	return &Cache{expiredDuration: expiredDuration, C: make(chan uint64, 10)}
}

func (c *Cache) Store(key interface{}, value interface{}) {
	c.Map.Store(key, value)
	if v, ok := value.(*CacheItme); ok {
		v.UpdateExpiredTime(c.expiredDuration)
	}
}

type CacheItme struct {
	expiredTime time.Time

	Object interface{}
}

func (i *CacheItme) IsExpired() bool {
	return i.expiredTime.Before(time.Now())
}

func (i *CacheItme) UpdateExpiredTime(duration time.Duration) {
	i.expiredTime = time.Now().Add(duration)
}

type memberCache struct {
	ProjectID uint64
	UserID    uint
	Name      string
	Nick      string
}

// projectClusterNamespaceCache caches the relationship for project:cluster:namespace
type projectClusterNamespaceCache struct {
	ProjectID  uint64
	Namespaces map[string][]string
}

func newProjectClusterNamespaceCache(projectID uint64) *projectClusterNamespaceCache {
	return &projectClusterNamespaceCache{
		ProjectID:  projectID,
		Namespaces: make(map[string][]string),
	}
}

type quotaItem struct {
	ClusterName string
	CPUQuota    uint64
	MemQuota    uint64
}
