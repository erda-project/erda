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

package meta

import (
	"github.com/go-redis/redis"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (p *provider) setCache() error {
	meta, ok := p.Meta.Load().([]MetricMeta)
	if !ok {
		return nil
	}
	expire := 3 * p.Cfg.ReloadInterval

	var caches []MetricMeta
	caches = append(caches, meta...)

	text, err := json.MarshalToString(caches)

	if err != nil {
		return err
	}
	if len(meta) <= 0 {
		return nil
	}
	err = p.Redis.Set(p.Cfg.CacheKeyPrefix, text, expire).Err()
	return err
}

func (p *provider) getCache() ([]MetricMeta, error) {
	var caches []MetricMeta

	val, err := p.Redis.Get(p.Cfg.CacheKeyPrefix).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if len(val) == 0 {
		return nil, nil
	}
	err = json.UnmarshalFromString(val, &caches)
	if err != nil {
		p.Log.Warnf("corrupted table cached-data: \n%s\n", val)
		return nil, err
	}
	return caches, nil
}
