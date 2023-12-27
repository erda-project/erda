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

package addon

import (
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

const (
	DefaultTtl  = 10 * time.Minute
	DefaultSize = 500
)

type Cache struct {
	gcache.Cache
	bdl *bundle.Bundle
}

// VersionMap Version of the corresponding addon
type VersionMap map[string]apistructs.ExtensionVersion

func (v *VersionMap) GetDefault() (apistructs.ExtensionVersion, bool) {
	res, ok := (*v)["default"]
	return res, ok
}

var (
	cache *Cache
	mutex = &sync.Mutex{}
)

// GetCache Get the Cache in unary mode, if the Cache is not initialized, then use the default configuration to initialize it.
func GetCache() *Cache {
	mutex.Lock()
	defer mutex.Unlock()
	if cache == nil {
		logrus.Infof("Addon cache is not initialized and will be initialized using the default configuration")
		InitCache(DefaultTtl, DefaultSize)
	}
	return cache
}

// InitCache Initialize the Cache
func InitCache(ttl time.Duration, size int) {
	// init Bundle
	bundleOpts := []bundle.Option{
		bundle.WithHTTPClient(
			httpclient.New(
				httpclient.WithTimeout(time.Second, time.Second*60),
			)),
		bundle.WithErdaServer(),
	}
	bdl := bundle.New(bundleOpts...)

	var addonCache = &Cache{
		bdl: bdl,
	}

	addonCache.initCache(ttl, size)

	cache = addonCache
}

// InitCache Initialize the gcache.Cache
func (a *Cache) initCache(ttl time.Duration, size int) {
	a.Cache = gcache.New(size).LRU().Expiration(ttl).LoaderFunc(func(key interface{}) (interface{}, error) {
		addonName := key.(string)
		addons, err := a.bdl.QueryExtensionVersions(apistructs.ExtensionVersionQueryRequest{Name: addonName, All: true})
		if err != nil {
			logrus.Errorf("failed to query addon: %v", addonName)
			return nil, err
		}
		versions := VersionMap{}
		for _, addon := range addons {
			versions[addon.Version] = addon
			if addon.IsDefault {
				versions["default"] = addon
			}
		}
		return &versions, nil
	}).Build()
}
