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

package indexloader

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

func (p *provider) runCacheLoader(ctx context.Context) error {
	p.Log.Infof("start cache-indices loader")
	defer p.Log.Info("exit cache-indices loader")
	timer := time.NewTimer(p.Cfg.IndexReloadInterval / 2)
	defer timer.Stop()
	var notifiers []chan error
	for {
		p.syncLock.Lock()
		for p.inSync {
			p.cond.Wait()
		}
		startSyncCh := p.startSyncCh

		select {
		case <-ctx.Done():
			p.syncLock.Unlock()
			return nil
		case ch := <-p.reloadCh:
			if ch != nil {
				notifiers = append(notifiers, ch)
			}
		case <-startSyncCh:
			p.inSync = false
			p.syncLock.Unlock()
			continue
		case <-timer.C:
		}

	drain:
		for {
			select {
			case ch := <-p.reloadCh:
				if ch != nil {
					notifiers = append(notifiers, ch)
				}
			default:
				break drain
			}
		}

		err := p.reloadIndicesFromCache(ctx)
		if err != nil {
			p.Log.Errorf("failed to reload indices: %s", err)
		}
		for _, n := range notifiers {
			n <- err
		}
		notifiers = nil
		timer.Reset(p.Cfg.IndexReloadInterval)
		p.syncLock.Unlock()
	}
}

func (p *provider) reloadIndicesFromCache(ctx context.Context) error {
	indices, err := p.queryIndexFromCache()
	if err != nil {
		return nil
	}
	ch := make(chan struct{})
	p.updateIndices(&indicesBundle{
		indices: indices,
		doneCh:  ch,
	})
	select {
	case <-ch:
	case <-ctx.Done():
		return nil
	}
	return nil
}

func (p *provider) storeIndicesToCache(indices map[string]*IndexGroup) error {
	expiration := 2*p.Cfg.IndexReloadInterval + p.Cfg.RequestTimeout
	byts, _ := json.Marshal(indices)
	if err := p.Redis.Set(p.Cfg.CacheKeyPrefix+"-all", string(byts), expiration).Err(); err != nil {
		return err
	}

	// requestDuration := time.Duration(int64(100*time.Millisecond) * int64(p.getIndicesNum(indices)))
	// expiration := 2*p.Cfg.IndexReloadInterval + requestDuration
	// marshal := func(i *IndexEntry) string {
	// 	byts, _ := json.Marshal(i)
	// 	return string(byts)
	// }
	// prefix := p.Cfg.CacheKeyPrefix + "-"
	// for _, index := range indices {
	// 	for _, ns := range index.Groups {
	// 		if ns.Fixed != nil {
	// 			if err := p.Redis.Set(prefix+ns.Fixed.Index, marshal(ns.Fixed), expiration).Err(); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		for _, item := range ns.List {
	// 			if err := p.Redis.Set(prefix+item.Index, marshal(item), expiration).Err(); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		for _, keys := range ns.Groups {
	// 			if keys.Fixed != nil {
	// 				if err := p.Redis.Set(prefix+keys.Fixed.Index, marshal(keys.Fixed), expiration).Err(); err != nil {
	// 					return err
	// 				}
	// 			}
	// 			for _, item := range keys.List {
	// 				if err := p.Redis.Set(prefix+item.Index, marshal(item), expiration).Err(); err != nil {
	// 					return err
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	return nil
}

func (p *provider) queryIndexFromCache() (map[string]*IndexGroup, error) {
	start := time.Now()
	p.Log.Debugf("start query indices from cache")
	indices := make(map[string]*IndexGroup)
	value, err := p.Redis.Get(p.Cfg.CacheKeyPrefix + "-all").Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if len(value) > 0 {
		err = json.NewDecoder(strings.NewReader(value)).Decode(&indices)
		if err != nil {
			p.Log.Warnf("corrupted index cached-data:\n%s\n", value)
		}
	}

	// slow query
	// prefix := p.Cfg.CacheKeyPrefix + "-"
	// keys, err := p.Redis.Keys(prefix + "*").Result()
	// if err != nil {
	// 	return nil, err
	// }
	// indices := make(map[string]*indexGroup)
	// for _, key := range keys {
	// 	value, err := p.Redis.Get(key).Result()
	// 	if err != nil {
	// 		if err == redis.Nil {
	// 			continue
	// 		}
	// 		return nil, err
	// 	}
	// 	if len(value) <= 0 {
	// 		continue
	// 	}
	// 	entry := &IndexEntry{}
	// 	err = json.NewDecoder(strings.NewReader(value)).Decode(entry)
	// 	if err != nil {
	// 		p.Log.Warnf("corrupted index cached-data: key=%q, value=%q", key, value)
	// 		continue
	// 	}
	// 	metricGroup := indices[entry.Metric]
	// 	if metricGroup == nil {
	// 		metricGroup = &indexGroup{
	// 			Groups: make(map[string]*indexGroup),
	// 		}
	// 		indices[entry.Metric] = metricGroup
	// 	}
	// 	nsGroup := metricGroup.Groups[entry.Namespace]
	// 	if nsGroup == nil {
	// 		nsGroup = &indexGroup{
	// 			Groups: make(map[string]*indexGroup),
	// 		}
	// 		metricGroup.Groups[entry.Namespace] = nsGroup
	// 	}
	// 	if len(entry.Key) > 0 {
	// 		keyGroup := nsGroup.Groups[entry.Key]
	// 		if keyGroup == nil {
	// 			keyGroup = &indexGroup{}
	// 			nsGroup.Groups[entry.Key] = keyGroup
	// 		}
	// 		if entry.Fixed {
	// 			keyGroup.Fixed = entry
	// 		} else {
	// 			keyGroup.List = append(keyGroup.List, entry)
	// 		}
	// 	} else {
	// 		if entry.Fixed {
	// 			nsGroup.Fixed = entry
	// 		} else {
	// 			nsGroup.List = append(nsGroup.List, entry)
	// 		}
	// 	}
	// }

	p.Log.Debugf("got indices from cache, indices %d, metrics: %d, duration: %s", p.getIndicesNum(indices), len(indices), time.Since(start))
	return indices, nil
}
