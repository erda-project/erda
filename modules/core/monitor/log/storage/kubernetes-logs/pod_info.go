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

package kuberneteslogs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bluele/gcache"
	"github.com/go-redis/redis"
	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/encoding/jsonx"

	"github.com/erda-project/erda/modules/core/monitor/log/storage"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
)

// PodInfoQueryer .
type PodInfoQueryer interface {
	GetPodInfo(cid string, sel *storage.Selector) (tags map[string]string, err error)
}

type podInfoQueryer struct {
	cache1 gcache.Cache
	cache2 *redis.Client
	loader loader.Interface `autowired:"elasticsearch.index.loader@log"`
}

const cacheKeyPrefix = "container-"

func newPodInfoQueryer(p *provider) PodInfoQueryer {
	q := &podInfoQueryer{
		cache2: p.Redis,
	}
	q.cache1 = gcache.New(p.Cfg.PodInfoCacheSize).LRU().LoaderFunc(func(key interface{}) (interface{}, error) {
		cid := key.(string)
		cacheKey := cacheKeyPrefix + cid
		result, err := q.cache2.Get(cacheKey).Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}
		if len(result) > 0 {
			m := make(map[string]string)
			err := json.NewDecoder(strings.NewReader(result)).Decode(&m)
			if err == nil && len(m) > 0 {
				return m, nil
			}
		}
		info, err := p.queryPodInfo(cid)
		if err != nil {
			return nil, err
		}
		if len(info) <= 0 {
			return nil, fmt.Errorf("got empty pod info")
		}
		sb := &strings.Builder{}
		json.NewEncoder(sb).Encode(info)
		_, err = q.cache2.SetNX(cacheKey, sb.String(), p.Cfg.PodInfoCacheExpiration).Result()
		if err != nil {
			p.Log.Errorf("failed to SetNX pod info to redis")
		}
		return info, nil
	}).Build()
	return q
}

func (q *podInfoQueryer) GetPodInfo(cid string, sel *storage.Selector) (tags map[string]string, err error) {
	val, err := q.cache1.Get(cid)
	if err != nil {
		return nil, err
	}
	m, _ := val.(map[string]string)
	if len(m) > 0 {
		return m, err
	}
	return nil, fmt.Errorf("not found pod info")
}

func (p *provider) queryPodInfo(cid string) (map[string]string, error) {
	indices := p.getQueryPodInfoIndices(cid)
	if len(indices) <= 0 {
		return nil, nil
	}

	client := p.Loader.Client()
	searchSource := elastic.NewSearchSource()
	query := elastic.NewBoolQuery().Filter(elastic.NewTermQuery("tags.container_id", cid))
	searchSource.Query(query).Size(1)

	ctx, cancel := context.WithTimeout(p.ctx, p.Loader.RequestTimeout())
	defer cancel()
	resp, err := client.Search(indices...).
		IgnoreUnavailable(true).AllowNoIndices(true).
		SearchSource(searchSource).Do(ctx)
	if err != nil || (resp != nil && resp.Error != nil) {
		if resp != nil && resp.Error != nil {
			return nil, fmt.Errorf("failed to request pod info for log query: %s", jsonx.MarshalAndIndent(resp.Error))
		}
		return nil, fmt.Errorf("failed to request pod info for log query: %s", err)
	}
	if resp.Hits == nil || len(resp.Hits.Hits) <= 0 || resp.Hits.Hits[0].Source == nil {
		return nil, nil
	}
	m := make(map[string]interface{})
	err = json.NewDecoder(bytes.NewReader([]byte(*resp.Hits.Hits[0].Source))).Decode(&m)
	if err != nil {
		return nil, fmt.Errorf("bad pod info: %s", err)
	}
	tagvals, ok := m["tags"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("not found log tags")
	}
	tags := make(map[string]string)
	for k, v := range tagvals {
		val, ok := v.(string)
		if ok {
			tags[k] = val
		}
	}
	return tags, nil
}

func (p *provider) getQueryPodInfoIndices(cid string) []string {
	now := time.Now()
	end := now.UnixNano()
	start := now.Add(-24 * time.Hour).UnixNano()
	return p.Loader.Indices(p.ctx, start, end, loader.KeyPath{
		Recursive: true,
	})
}
