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

package indexmanager

import (
	"context"
	"fmt"
	"time"

	indexloader "github.com/erda-project/erda/modules/core/monitor/metric/index-loader"
	"github.com/olivere/elastic"
)

func (p *provider) runCleanIndices(ctx context.Context) {
	p.Loader.WaitAndGetIndices(ctx)
	p.Log.Infof("enable indices clean with interval(%v)", p.Cfg.IndexCleanInterval)
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			err := p.CleanIndices(ctx, func(*indexloader.IndexEntry) bool { return true })
			if err != nil {
				p.Log.Errorf("failed to CleanIndices: %s", err)
			}
		case req := <-p.clearCh:
			p.deleteIndices(req.list)
			if req.waitCh != nil {
				close(req.waitCh)
			}
		case <-ctx.Done():
			return
		}
		timer.Reset(p.Cfg.IndexCleanInterval)
	}
}

// CleanIndices .
func (p *provider) CleanIndices(ctx context.Context, filter IndexMatcher) error {
	indices := p.Loader.AllIndices()
	if len(indices) <= 0 {
		return nil
	}
	mc := p.getMetricConfig(ctx)
	select {
	case <-ctx.Done():
		return nil
	default:
	}
	now := time.Now()
	var removeList []string
	for _, mg := range indices {
		for _, ng := range mg.Groups {
			for _, entry := range ng.List {
				if filter(entry) && p.needToDelete(entry, mc, now) {
					// atomic.StoreInt32(&entry.Deleted, 1)
					removeList = append(removeList, entry.Index)
				}
			}
			for _, kg := range ng.Groups {
				for _, entry := range kg.List {
					if filter(entry) && p.needToDelete(entry, mc, now) {
						// atomic.StoreInt32(&entry.Deleted, 1)
						removeList = append(removeList, entry.Index)
					}
				}
			}
		}
	}
	if len(removeList) > 0 {
		err := p.deleteIndices(removeList)
		if err != nil {
			return err
		}
		p.Loader.ReloadIndices()
	}
	return nil
}

func (p *provider) needToDelete(entry *indexloader.IndexEntry, mc *metricConfig, now time.Time) bool {
	if entry.MaxT.IsZero() || (entry.Num > 0 && entry.Active) {
		return false
	}
	if len(entry.Key) > 0 {
		if mc != nil {
			if d, ok := mc.keysTTL[entry.Key]; ok {
				if int64(d) <= 0 {
					return false
				}
				return now.After(entry.MaxT.Add(d))
			}
		} else {
			return false
		}
	}
	if int64(p.Cfg.IndexTTL) <= 0 {
		return false
	}
	return now.After(entry.MaxT.Add(p.Cfg.IndexTTL))
}

func (p *provider) deleteIndices(removeList []string) error {
	const size = 10 // delete too much at once and the request will be rejected
	for len(removeList) >= size {
		err := p.deleteIndex(removeList[:size])
		if err != nil {
			return err
		}
		removeList = removeList[size:]
	}
	if len(removeList) > 0 {
		err := p.deleteIndex(removeList)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *provider) deleteIndex(indices []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.Cfg.RequestTimeout)
	defer cancel()
	resp, err := p.ES.Client().DeleteIndex(indices...).Do(ctx)
	if err != nil {
		if e, ok := err.(*elastic.Error); ok {
			if e.Status == 404 {
				return nil
			}
		}
		return err
	}
	if !resp.Acknowledged {
		return fmt.Errorf("delete indices Acknowledged=false")
	}
	p.Log.Infof("clean indices %d, %v", len(indices), indices)
	return nil
}
