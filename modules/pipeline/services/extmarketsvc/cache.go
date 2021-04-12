// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package extmarketsvc

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

func InitializeCaches(svc *ExtMarketSvc) {
	extCaches = &cache{
		Extensions: make(map[string]*apistructs.ExtensionVersion),
		svc:        svc,
	}
	go extCaches.continueUpdate()
}

type cache struct {
	Extensions map[string]*apistructs.ExtensionVersion // key: name@version, value: ext

	svc  *ExtMarketSvc
	lock sync.Mutex
}

var extCaches *cache

func (c *cache) getExt(item string) *apistructs.ExtensionVersion {
	c.lock.Lock()
	defer c.lock.Unlock()

	ext, ok := c.Extensions[item]
	if !ok {
		return nil
	}

	return ext
}

func (c *cache) updateExt(item string, ext apistructs.ExtensionVersion) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Extensions[item] = &ext
}

func (c *cache) updateCachedExts() {
	c.lock.Lock()
	defer c.lock.Unlock()

	// construct search request
	var existItems []string
	for item := range c.Extensions {
		existItems = append(existItems, item)
	}
	searchedActions, err := c.svc.doSearchRemoteActions(existItems)
	if err != nil {
		logrus.Errorf("failed to update cached extensions, items: %v, err: %v",
			strutil.Join(existItems, ", "), err)
	}

	// update to caches
	for item, ext := range searchedActions {
		c.Extensions[item] = &ext
	}
}

// continueUpdate update caches periodically.
func (c *cache) continueUpdate() {
	ticker := time.NewTicker(time.Minute * 5)
	for {
		select {
		case <-ticker.C:
			c.updateCachedExts()
		}
	}
}
