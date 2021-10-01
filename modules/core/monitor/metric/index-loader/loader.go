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

package indexloader

import (
	"context"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/encoding/jsonx"
	"github.com/recallsong/go-utils/lang/size"
)

// IndexEntry .
type IndexEntry struct {
	Index       string
	Metric      string
	Namespace   string
	Key         string
	Fixed       bool
	Num         int64
	MinT        time.Time
	MaxT        time.Time
	DocsCount   int
	DocsDeleted int
	StoreBytes  int64
	StoreSize   string
	Active      bool
}

// IndexEntrys .
type IndexEntrys []*IndexEntry

func (entrys IndexEntrys) Len() int      { return len(entrys) }
func (entrys IndexEntrys) Swap(i, j int) { entrys[i], entrys[j] = entrys[j], entrys[i] }
func (entrys IndexEntrys) Less(i, j int) bool {
	if entrys[i].Num == entrys[j].Num {
		if entrys[i].MinT.Equal(entrys[j].MinT) {
			if entrys[i].MaxT.Equal(entrys[j].MaxT) {
				return entrys[i].StoreBytes < entrys[j].StoreBytes
			}
			return entrys[i].MaxT.Before(entrys[j].MaxT)
		}
		return entrys[i].MinT.Before(entrys[j].MinT)
	}
	return entrys[i].Num < entrys[j].Num
}

// IndexGroup .
type IndexGroup struct {
	Groups map[string]*IndexGroup `json:"groups,omitempty"`
	List   []*IndexEntry          `json:"list,omitempty"`
	Fixed  *IndexEntry            `json:"fixed,omitempty"`
}

type timeRange struct {
	MinT        time.Time
	MaxT        time.Time
	DocsCount   int
	DocsDeleted int
}

func (p *provider) getIndicesFromES() (indices map[string]*IndexGroup, err error) {
	start := time.Now()
	p.Log.Debugf("start query indices from es")
	ctx, cancel := context.WithTimeout(context.Background(), p.Cfg.RequestTimeout)
	resps, err := p.ES.Client().CatIndices().Index(p.Cfg.IndexPrefix+"-*").Columns("index", "docs.count", "docs.deleted", "store.size").Do(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	cancel()
	indices = make(map[string]*IndexGroup)
	for _, item := range resps {
		parts := strings.Split(item.Index, "-")
		if len(parts) == 2 {
			// spot-empty
			continue
		}
		storeBytes, err := size.ParseBytes(item.StoreSize)
		if err != nil {
			storeBytes = -1
		}
		var entry *IndexEntry
		if len(parts) == 3 {
			// spot-<metric>-<namespace>
			// spot-<metric>-<namespace>.<key>
			entry = &IndexEntry{
				Index:       item.Index,
				Metric:      parts[1],
				Namespace:   parts[2],
				Num:         -1,
				Fixed:       true,
				DocsCount:   item.DocsCount,
				DocsDeleted: item.DocsDeleted,
				StoreBytes:  storeBytes,
				StoreSize:   item.StoreSize,
				Active:      true,
			}
		} else if len(parts) == 5 && parts[3] == "r" {
			// spot-<metric>-<namespace>-r-000001
			// spot-<metric>-<namespace>.<key>-r-000001
			n, err := strconv.ParseInt(parts[4], 10, 64)
			if err == nil {
				entry = &IndexEntry{
					Index:       item.Index,
					Metric:      parts[1],
					Namespace:   parts[2],
					Num:         n,
					DocsCount:   item.DocsCount,
					DocsDeleted: item.DocsDeleted,
					StoreBytes:  storeBytes,
					StoreSize:   item.StoreSize,
				}
			}
		} else if len(parts) == 5 && parts[3] == "m" {
			// spot-<metric>-<namespace>-m-xxx
			// spot-<metric>-<namespace>.<key>-m-xxx
			if err == nil {
				entry = &IndexEntry{
					Index:       item.Index,
					Metric:      parts[1],
					Namespace:   parts[2],
					Num:         0,
					DocsCount:   item.DocsCount,
					DocsDeleted: item.DocsDeleted,
					StoreBytes:  storeBytes,
					StoreSize:   item.StoreSize,
				}
			}
		} else if len(parts) == 4 {
			// spot-<metric>-<namespace>-<timestamp>
			// spot-<metric>-<namespace>.<key>-<timestamp>
			t, err := strconv.ParseInt(parts[3], 10, 64)
			if err == nil {
				mint := time.Unix(t/1000, (t%1000)*int64(time.Millisecond))
				maxt := mint.Add(24*time.Hour - time.Nanosecond)
				entry = &IndexEntry{
					Index:       item.Index,
					Metric:      parts[1],
					Namespace:   parts[2],
					Num:         -1,
					DocsCount:   item.DocsCount,
					DocsDeleted: item.DocsDeleted,
					StoreBytes:  storeBytes,
					StoreSize:   item.StoreSize,
					MinT:        mint,
					MaxT:        maxt,
				}
			}
		}
		if entry == nil {
			p.Log.Debugf("invalid index format %s", item.Index)
			continue
		}
		idx := strings.Index(entry.Namespace, ".")
		if idx >= 0 {
			entry.Key = entry.Namespace[idx+1:]
			entry.Namespace = entry.Namespace[0:idx]
		}
		metricGroup := indices[entry.Metric]
		if metricGroup == nil {
			metricGroup = &IndexGroup{
				Groups: make(map[string]*IndexGroup),
			}
			indices[entry.Metric] = metricGroup
		}

		nsGroup := metricGroup.Groups[entry.Namespace]
		if nsGroup == nil {
			nsGroup = &IndexGroup{
				Groups: make(map[string]*IndexGroup),
			}
			metricGroup.Groups[entry.Namespace] = nsGroup
		}
		if len(entry.Key) > 0 {
			keyGroup := nsGroup.Groups[entry.Key]
			if keyGroup == nil {
				keyGroup = &IndexGroup{}
				nsGroup.Groups[entry.Key] = keyGroup
			}
			if entry.Fixed {
				keyGroup.Fixed = entry
			} else {
				keyGroup.List = append(keyGroup.List, entry)
			}
		} else {
			if entry.Fixed {
				nsGroup.Fixed = entry
			} else {
				nsGroup.List = append(nsGroup.List, entry)
			}
		}
	}
	p.Log.Debugf("got indices from es, indices %d, metrics: %d, duration: %s", p.getIndicesNum(indices), len(indices), time.Since(start))
	return indices, nil
}

func (p *provider) getIndicesFromESWithTimeRange() (map[string]*IndexGroup, error) {
	indices, err := p.getIndicesFromES()
	if err != nil {
		return nil, err
	}
	// the maximum and minimum times to query each index
	for _, index := range indices {
		for _, ns := range index.Groups {
			p.initIndexGroup(ns)
			for _, keys := range ns.Groups {
				p.initIndexGroup(keys)
			}
		}
	}
	p.cleanTimeRangeCache(indices)
	return indices, nil
}

func (p *provider) reloadIndicesFromES(ctx context.Context) error {
	indices, err := p.getIndicesFromESWithTimeRange()
	if err != nil {
		return err
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

// func (p *provider) reloadIndicesFromES() error {
// 	index, num, err := p.getIndicesFromESWithoutTime()

// 	// // the maximum and minimum times to query each index
// 	// for _, index := range indices {
// 	// 	for _, ns := range index.Groups {
// 	// 		m.initIndexGroup(ns)
// 	// 		for _, keys := range ns.Groups {
// 	// 			m.initIndexGroup(keys)
// 	// 		}
// 	// 	}
// 	// }
// 	// m.cleanTimeRangeCache(indices)

// 	// m.indices.Store(indices)
// 	// m.log.Infof("load indices %d, metrics: %d", indexNum, len(indices))

// 	// m.createdLock.Lock()
// 	// if len(m.created) > 0 {
// 	// 	m.created = make(map[string]bool)
// 	// }
// 	// m.createdLock.Unlock()
// 	return nil
// }

func (p *provider) initIndexGroup(ig *IndexGroup) {
	var maxn, maxt int64 = math.MinInt64, math.MinInt64
	var maxNumEntry, maxTimeEntry *IndexEntry
	for _, item := range ig.List {
		if item.Num < 0 {
			t := item.MinT.UnixNano()
			if t >= maxt {
				maxt = t
				maxTimeEntry = item
			}
		} else if item.Num > 0 {
			if item.Num >= maxn {
				maxn = item.Num
				maxNumEntry = item
			}
		}
	}
	if maxNumEntry != nil {
		maxNumEntry.Active = true
	} else if maxTimeEntry != nil {
		maxTimeEntry.Active = true
	}
	for _, entry := range ig.List {
		p.setupTimeRange(entry)
	}
	sort.Sort(sort.Reverse(IndexEntrys(ig.List)))
}

func (p *provider) setupTimeRange(index *IndexEntry) {
	if p.Cfg.QueryIndexTimeRange && !index.Active && index.Num >= 0 {
		ranges, ok := p.timeRanges[index.Index]
		if !ok || (index.DocsCount != ranges.DocsCount || index.DocsDeleted != ranges.DocsDeleted) {
			searchSource := elastic.NewSearchSource()
			searchSource.Aggregation("min_time", elastic.NewMinAggregation().Field("timestamp"))
			searchSource.Aggregation("max_time", elastic.NewMaxAggregation().Field("timestamp"))
			context, cancel := context.WithTimeout(context.Background(), p.Cfg.RequestTimeout)
			defer cancel()
			resp, err := p.ES.Client().Search(index.Index).IgnoreUnavailable(true).AllowNoIndices(true).
				SearchSource(searchSource).Do(context)
			if err != nil {
				p.Log.Errorf("failed to query index %q time range: %s", index.Index, err)
				return
			} else if resp != nil && resp.Error != nil {
				p.Log.Errorf("failed to query index %q time range: %s", index.Index, jsonx.MarshalAndIndent(resp.Error))
				return
			}
			min, ok := resp.Aggregations.Min("min_time")
			if ok && min.Value != nil {
				t := int64(*min.Value)
				index.MinT = time.Unix(t/int64(time.Second), t%int64(time.Second))
			}
			max, ok := resp.Aggregations.Max("max_time")
			if ok && max.Value != nil {
				t := int64(*max.Value)
				index.MaxT = time.Unix(t/int64(time.Second), t%int64(time.Second))
			}
			p.Log.Debugf("query index %q , mint: %q, maxt: %q", index.Index, index.MinT.String(), index.MaxT.String())
			if min != nil && min.Value != nil &&
				max != nil && max.Value != nil {
				p.timeRanges[index.Index] = &timeRange{
					MinT:        index.MinT,
					MaxT:        index.MaxT,
					DocsCount:   index.DocsCount,
					DocsDeleted: index.DocsDeleted,
				}
			}
		} else {
			index.MinT = ranges.MinT
			index.MaxT = ranges.MaxT
		}
	}
}

func (p *provider) cleanTimeRangeCache(indices map[string]*IndexGroup) {
	set := make(map[string]bool)
	for _, index := range indices {
		for _, ns := range index.Groups {
			for _, entry := range ns.List {
				set[entry.Index] = true
			}
			for _, keys := range ns.Groups {
				for _, entry := range keys.List {
					set[entry.Index] = true
				}
			}
		}
	}
	for index := range p.timeRanges {
		if !set[index] {
			delete(p.timeRanges, index)
		}
	}
}

func (p *provider) syncIndiceToCache(ctx context.Context) {
	p.Log.Infof("start indices-cache sycn task")
	defer p.Log.Info("exit indices-cache sycn task")

	if p.startSyncCh != nil {
		close(p.startSyncCh)
	}
	p.syncLock.Lock()
	p.inSync = true
	p.startSyncCh = make(chan struct{})
	defer func() {
		p.inSync = false
		p.syncLock.Unlock()
		p.cond.Broadcast()
	}()

	timer := time.NewTimer(p.Cfg.IndexReloadInterval / 2)
	defer timer.Stop()
	var notifiers []chan error
	for {
		select {
		case <-ctx.Done():
			return
		case ch := <-p.reloadCh:
			if ch != nil {
				notifiers = append(notifiers, ch)
			}
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

		err := p.doSyncIndiceToCache(ctx)
		if err != nil {
			p.Log.Errorf("failed to sync indices: %s", err)
		}
		for _, n := range notifiers {
			n <- err
			close(n)
		}
		notifiers = nil
		timer.Reset(p.Cfg.IndexReloadInterval)
	}
}

func (p *provider) doSyncIndiceToCache(ctx context.Context) error {
	indices, err := p.getIndicesFromESWithTimeRange()
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return context.Canceled
	case p.setIndicesCh <- &indicesBundle{indices: indices}:
	}
	return p.storeIndicesToCache(indices)
}

func (p *provider) runElasticSearchIndexLoader(ctx context.Context) error {
	p.Log.Infof("start elasticsearch-indices loader")
	defer p.Log.Info("exit elasticsearch-indices loader")
	timer := time.NewTimer(p.Cfg.IndexReloadInterval / 2)
	defer timer.Stop()
	var notifiers []chan error
	for {
		select {
		case <-ctx.Done():
			return nil
		case ch := <-p.reloadCh:
			if ch != nil {
				notifiers = append(notifiers, ch)
			}
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

		err := p.reloadIndicesFromES(ctx)
		if err != nil {
			p.Log.Errorf("failed to reload indices: %s", err)
		}
		for _, n := range notifiers {
			n <- err
			close(n)
		}
		notifiers = nil
		timer.Reset(p.Cfg.IndexReloadInterval)
	}
}
