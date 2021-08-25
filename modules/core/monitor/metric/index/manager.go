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
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/encoding/jsonx"
	"github.com/recallsong/go-utils/lang/size"

	"github.com/erda-project/erda-infra/base/logs"
	mutex "github.com/erda-project/erda-infra/providers/etcd-mutex"
	"github.com/erda-project/erda/modules/core/monitor/metric"
	"github.com/erda-project/erda/pkg/router"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

// IndexManager loading indexes, managing index creation, scrolling, etc
type IndexManager struct {
	client                   *elastic.Client
	db                       *gorm.DB
	cfg                      *config
	urls                     string
	rolloverBody             string
	rolloverBodyForDiskClean string
	log                      logs.Logger

	indices    atomic.Value          // map[string]*indexGroup, the index that has been loaded
	reloadCh   chan chan struct{}    // Actively trigger index loading
	timeRanges map[string]*timeRange // Cache the maximum and minimum time index query results without having to look them up every time the index is loaded

	iconfig          atomic.Value // Index setting
	namespaces       *router.Router
	defaultNamespace string

	created     map[string]bool // When alias:true, index&alias that was created but may not have been loaded in the indices
	createdLock sync.Mutex      // Locks access to created

	clearCh chan *clearRequest
	closeCh chan struct{}
}

type indexGroup struct {
	Groups map[string]*indexGroup `json:"groups,omitempty"`
	List   []*IndexEntry          `json:"list,omitempty"`
	Fixed  *IndexEntry            `json:"fixed,omitempty"`
}

type timeRange struct {
	MinT        time.Time
	MaxT        time.Time
	DocsCount   int
	DocsDeleted int
}

// NewIndexManager .
func NewIndexManager(cfg *config, es *elastic.Client, urls string, db *gorm.DB, log logs.Logger) *IndexManager {
	r := router.New()
	for _, item := range cfg.Namespaces {
		if len(item.Tags) == 1 && len(item.Namespace) == 0 {
			item.Namespace = item.Tags[0].Value
		}
		if len(item.Tags) <= 0 || len(item.Namespace) == 0 {
			continue
		}
		for _, name := range strings.Split(item.Name, ",") {
			r.Add(name, item.Tags, normalizeIndexPart(item.Namespace))
		}
	}
	r.PrintTree(true)
	return &IndexManager{
		client:           es,
		db:               db,
		cfg:              cfg,
		urls:             urls,
		namespaces:       r,
		defaultNamespace: normalizeIndexPart(cfg.DefaultNamespace),
		log:              log,
		reloadCh:         make(chan chan struct{}),
		timeRanges:       make(map[string]*timeRange),
		created:          make(map[string]bool),
		clearCh:          make(chan *clearRequest),
		closeCh:          make(chan struct{}),
	}
}

const timeForSplitIndex int64 = 24 * int64(time.Hour)

// GetWriteIndex .
// enable rollover:
// spot-<metric>-<namespace>-r-000001
// spot-<metric>-<namespace>.<key>-r-000001
// unable rollover:
// spot-<metric>-<namespace>-<timestamp>
// spot-<metric>-<namespace>.<key>-<timestamp>
func (m *IndexManager) GetWriteIndex(metric *metric.Metric) (string, bool) {
	ns, key := m.getNamespace(metric), m.getKey(metric)
	suffix := m.getIndexSuffix(ns, key)
	name := normalizeIndexPart(strings.ToLower(metric.Name))

	if m.cfg.EnableRollover {
		alias := m.indexAlias(name, suffix)
		// Check if the index exists
		indices := m.waitAndGetIndices()
		var find bool
		metricGroup, ok := indices[name]
		if ok {
			nsGroup, ok := metricGroup.Groups[ns]
			if ok {
				if len(key) > 0 {
					keysGroup, ok := nsGroup.Groups[key]
					if ok {
						if len(keysGroup.List) > 0 && keysGroup.List[0].Num > 0 {
							find = true
						}
					}
				} else {
					if len(nsGroup.List) > 0 && nsGroup.List[0].Num > 0 {
						find = true
					}
				}
			}
		}
		return alias, find
	}
	timestamp := (metric.Timestamp - metric.Timestamp%timeForSplitIndex) / 1000000
	return m.cfg.IndexPrefix + "-" + name + "-" + suffix + "-" + strconv.FormatInt(timestamp, 10), true
}

// GetWriteFixedIndex spot-<metric>-<namespace> Gets a time-free index, which represents permanent storage of data
func (m *IndexManager) GetWriteFixedIndex(metric *metric.Metric) string {
	return m.cfg.IndexPrefix + "-" + normalizeIndexPart(strings.ToLower(metric.Name)) + "-" +
		m.getIndexSuffix(m.getNamespace(metric), m.getKey(metric))
}

func normalizeIndexPart(s string) string {
	return strings.Replace(s, "-", "_", -1)
}

func (m *IndexManager) indexAlias(name, suffix string) string {
	return m.cfg.IndexPrefix + "-" + name + "-" + suffix + "-rollover"
}

func (m *IndexManager) getNamespace(metric *metric.Metric) string {
	ns := m.namespaces.Find(metric.Name, metric.Tags)
	if ns != nil {
		return ns.(string)
	}
	return m.defaultNamespace
}

func (m *IndexManager) getIndexSuffix(ns, key string) string {
	if len(key) > 0 {
		return ns + "." + key
	}
	return ns
}

// CreateIndex .
func (m *IndexManager) CreateIndex(metric *metric.Metric) error {
	ns, key := m.getNamespace(metric), m.getKey(metric)
	suffix := m.getIndexSuffix(ns, key)
	name := normalizeIndexPart(strings.ToLower(metric.Name))
	alias := m.indexAlias(name, suffix)

	m.createdLock.Lock()
	defer m.createdLock.Unlock()
	if m.created[alias] {
		return nil // 索引已经创建
	}
	index := m.cfg.IndexPrefix + "-" + name + "-" + suffix + "-r-000001"
	err := m.createIndexWithRetry(
		index, // first index
		alias,
	)
	if err != nil {
		m.log.Error(err)
		return err
	}
	m.created[alias] = true
	m.log.Infof("create index %q with alias %q ok", index, alias)
	return nil
}

func (m *IndexManager) createIndexWithRetry(index, alias string) (err error) {
	createIndex := func(index, alias string) (*elastic.IndicesCreateResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.RequestTimeout)
		defer cancel()
		return m.client.CreateIndex(index).BodyJson(
			map[string]interface{}{
				"aliases": map[string]interface{}{
					alias: make(map[string]interface{}),
				},
			},
		).Do(ctx)
	}
	for i := 0; i < 2; i++ {
		resp, e := createIndex(index, alias)
		if e == nil {
			if resp != nil && !resp.Acknowledged {
				return fmt.Errorf("fail to create index %q, alias %q: not Acknowledged", index, alias)
			}
			return nil
		}
		err = e
	}
	return fmt.Errorf("fail to create index %q, alias %q: %s", index, alias, err)
}

// Start .
func (m *IndexManager) Start(lock mutex.Mutex) error {
	// Load the configuration
	if m.cfg.LoadIndexTTLFromDatabase {
		if int64(m.cfg.TTLReloadInterval) <= 0 {
			return fmt.Errorf("invalid TTLReloadInterval: %v", m.cfg.TTLReloadInterval)
		}
		go func() {
			m.log.Infof("enable indices ttl reload from database, interval: %v", m.cfg.TTLReloadInterval)
			tick := time.Tick(m.cfg.TTLReloadInterval)
			for {
				m.loadConfig()
				select {
				case <-tick:
				case <-m.closeCh:
					return
				}
			}
		}()
	}
	if m.cfg.EnableIndexClean {
		err := m.startClean()
		if err != nil {
			return err
		}
		if m.cfg.DiskClean.EnableIndexCleanByDisk {
			err = m.startDiskCheck(lock)
			if err != nil {
				return err
			}
		}
	}
	// index rollover
	if m.cfg.EnableRollover {
		err := m.startRollover()
		if err != nil {
			return err
		}
	}

	// Load index task
	m.log.Infof("start indices reload, interval: %v", m.cfg.IndexReloadInterval)
	tick := time.Tick(m.cfg.IndexReloadInterval)
	var done chan struct{}
	for {
		err := m.reloadIndices()
		if err != nil {
			m.log.Errorf("fail to reload indices: %s", err)
		}
		if done != nil {
			close(done)
			done = nil
		}
		// Timeout or active trigger to load the index
		select {
		case ch, ok := <-m.reloadCh:
			if !ok {
				return nil
			}
			done = ch
		case <-tick:
		case <-m.closeCh:
			return nil
		}
	}
}

// reload
func (m *IndexManager) toReloadIndices(wait bool) {
	if wait {
		ch := make(chan struct{})
		m.reloadCh <- ch
		<-ch
	} else {
		m.reloadCh <- nil
	}
}

// Close .
func (m *IndexManager) Close() error {
	// close(m.reloadCh)
	close(m.closeCh)
	return nil
}

// WaitIndicesLoad .
func (m *IndexManager) WaitIndicesLoad() {
	m.waitAndGetIndices()
}

func (m *IndexManager) waitAndGetIndices() map[string]*indexGroup {
	for {
		v := m.indices.Load()
		if v == nil {
			// Wait for the index to complete loading
			time.Sleep(1 * time.Second)
			continue
		}
		return v.(map[string]*indexGroup)
	}
}

func (m *IndexManager) reloadIndices() error {
	ctx, cancel := context.WithTimeout(context.Background(), m.cfg.RequestTimeout)
	resps, err := m.client.CatIndices().Index(m.cfg.IndexPrefix+"-*").Columns("index", "docs.count", "docs.deleted", "store.size").Do(ctx)
	if err != nil {
		cancel()
		return err
	}
	cancel()
	var indexNum int
	indices := make(map[string]*indexGroup)
	for _, item := range resps {
		// spot-<metric>-<namespace>-<timestamp>
		// spot-<metric>-<namespace>.<key>-<timestamp>
		// spot-<metric>-<namespace>-r-000001
		// spot-<metric>-<namespace>.<key>-r-000001
		// spot-<metric>-<namespace>
		// spot-<metric>-<namespace>.<key>
		// spot-empty
		parts := strings.Split(item.Index, "-")
		if len(parts) == 2 {
			continue
		}
		storeSize, err := size.ParseBytes(item.StoreSize)
		if err != nil {
			storeSize = -1
		}
		var entry *IndexEntry
		if len(parts) == 3 {
			entry = &IndexEntry{
				Index:       item.Index,
				Metric:      parts[1],
				Namespace:   parts[2],
				Num:         -1,
				Fixed:       true,
				DocsCount:   item.DocsCount,
				DocsDeleted: item.DocsDeleted,
				StoreSize:   storeSize,
				Active:      true,
			}
		} else if len(parts) == 5 && parts[3] == "r" {
			n, err := strconv.ParseInt(parts[4], 10, 64)
			if err == nil {
				entry = &IndexEntry{
					Index:       item.Index,
					Metric:      parts[1],
					Namespace:   parts[2],
					Num:         n,
					DocsCount:   item.DocsCount,
					DocsDeleted: item.DocsDeleted,
					StoreSize:   storeSize,
				}
			}
		} else if len(parts) == 5 && parts[3] == "m" {
			if err == nil {
				entry = &IndexEntry{
					Index:       item.Index,
					Metric:      parts[1],
					Namespace:   parts[2],
					Num:         0,
					DocsCount:   item.DocsCount,
					DocsDeleted: item.DocsDeleted,
					StoreSize:   storeSize,
				}
			}
		} else if len(parts) == 4 {
			// previous index
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
					StoreSize:   storeSize,
					MinT:        mint,
					MaxT:        maxt,
				}
			}
		}
		if entry == nil {
			m.log.Debugf("invalid index format %s", item.Index)
			continue
		}
		idx := strings.Index(entry.Namespace, ".")
		if idx >= 0 {
			entry.Key = entry.Namespace[idx+1:]
			entry.Namespace = entry.Namespace[0:idx]
		}
		metricGroup := indices[entry.Metric]
		if metricGroup == nil {
			metricGroup = &indexGroup{
				Groups: make(map[string]*indexGroup),
			}
			indices[entry.Metric] = metricGroup
		}

		nsGroup := metricGroup.Groups[entry.Namespace]
		if nsGroup == nil {
			nsGroup = &indexGroup{
				Groups: make(map[string]*indexGroup),
			}
			metricGroup.Groups[entry.Namespace] = nsGroup
		}
		if len(entry.Key) > 0 {
			keyGroup := nsGroup.Groups[entry.Key]
			if keyGroup == nil {
				keyGroup = &indexGroup{}
				nsGroup.Groups[entry.Key] = keyGroup
			}
			// Save the index with the key
			if entry.Fixed {
				keyGroup.Fixed = entry
			} else {
				keyGroup.List = append(keyGroup.List, entry)
			}
		} else {
			// Save the index without the key
			if entry.Fixed {
				nsGroup.Fixed = entry
			} else {
				nsGroup.List = append(nsGroup.List, entry)
			}
		}

		indexNum++
	}

	// The maximum and minimum times to query each index
	for _, index := range indices {
		for _, ns := range index.Groups {
			m.initIndexGroup(ns)
			for _, keys := range ns.Groups {
				m.initIndexGroup(keys)
			}
		}
	}
	m.cleanTimeRangeCache(indices)

	// fmt.Println(jsonx.MarshalAndIndent(indices))
	m.indices.Store(indices)
	m.log.Infof("load indices %d, metrics: %d", indexNum, len(indices))

	// The index is already loaded, empty created
	m.createdLock.Lock()
	if len(m.created) > 0 {
		m.created = make(map[string]bool)
	}
	m.createdLock.Unlock()
	return nil
}

func (m *IndexManager) initIndexGroup(ig *indexGroup) {
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
	if m.cfg.EnableRollover {
		if maxNumEntry != nil {
			maxNumEntry.Active = true
		} else if maxTimeEntry != nil {
			maxTimeEntry.Active = true
		}
	} else {
		if maxNumEntry != nil {
			maxNumEntry.Active = true
		}
		if maxTimeEntry != nil {
			maxTimeEntry.Active = true
		}
	}
	for _, entry := range ig.List {
		m.setupTimeRange(entry)
	}
	sort.Sort(sort.Reverse(IndexEntrys(ig.List)))
}

// The maximum and minimum time to query the index
func (m *IndexManager) setupTimeRange(index *IndexEntry) {
	if m.cfg.QueryIndexTimeRange && !index.Active && index.Num >= 0 {
		ranges, ok := m.timeRanges[index.Index]
		// If the index has not been queried for a time range, or if the number of indexes has changed compared to the previous time range, the time range is re-queried
		if !ok || (index.DocsCount != ranges.DocsCount || index.DocsDeleted != ranges.DocsDeleted) {
			searchSource := elastic.NewSearchSource()
			searchSource.Aggregation("min_time", elastic.NewMinAggregation().Field("timestamp"))
			searchSource.Aggregation("max_time", elastic.NewMaxAggregation().Field("timestamp"))
			context, cancel := context.WithTimeout(context.Background(), m.cfg.RequestTimeout)
			defer cancel()
			resp, err := m.client.Search(index.Index).IgnoreUnavailable(true).AllowNoIndices(true).
				SearchSource(searchSource).Do(context)
			if err != nil {
				m.log.Errorf("fail to query index %q time range: %s", index.Index, err)
				return
			} else if resp != nil && resp.Error != nil {
				m.log.Errorf("fail to query index %q time range: %s", index.Index, jsonx.MarshalAndIndent(resp.Error))
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
			m.log.Debugf("query index %q , mint: %q, maxt: %q", index.Index, index.MinT.String(), index.MaxT.String())
			if min != nil && min.Value != nil &&
				max != nil && max.Value != nil {
				m.timeRanges[index.Index] = &timeRange{
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

func (m *IndexManager) cleanTimeRangeCache(indices map[string]*indexGroup) {
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
	// Clears the time-range cache for non-existent indexes
	for index := range m.timeRanges {
		if !set[index] {
			delete(m.timeRanges, index)
		}
	}
}

func matchTimeRange(entry *IndexEntry, start, end time.Time) bool {
	if (entry.MinT.IsZero() || entry.MinT.Before(end) || entry.MinT.Equal(end)) &&
		(entry.MaxT.IsZero() || entry.MaxT.After(start) || entry.MaxT.Equal(start)) {
		// if atomic.LoadInt32(&entry.Deleted) > 0 {
		// 	return false
		// }
		return true
	}
	return false
}

// GetReadIndices clusters as namespaces, start、end is ms
func (m *IndexManager) GetReadIndices(metrics []string, namespaces []string, start, end int64) (list []string) {
	if len(metrics) <= 0 {
		return []string{m.EmptyIndex()}
	}
	v := m.indices.Load()
	if v == nil {
		for _, name := range metrics {
			list = append(list, m.cfg.IndexPrefix+"-"+normalizeIndexPart(strings.ToLower(name))+"-*")
		}
	} else {
		indices := v.(map[string]*indexGroup)
		startT := time.Unix(start/1000, (start%1000)*int64(time.Millisecond))
		endT := time.Unix(end/1000, (end%1000)*int64(time.Millisecond)+999999) // It's down to nanoseconds, so plus 999999 is the last nanosecond of end
		for _, name := range metrics {
			name = normalizeIndexPart(strings.ToLower(name))
			ns, ok := indices[name]
			if !ok {
				continue
			}
			if len(namespaces) == 0 {
				for _, namespace := range ns.Groups {
					m.findIndex(namespace, startT, endT, &list)
				}
			} else {
				var appendDefaultNS bool
				for _, n := range namespaces {
					n = normalizeIndexPart(n)
					if n == m.defaultNamespace {
						appendDefaultNS = true
						continue
					}
					namespace, ok := ns.Groups[n]
					if ok {
						m.findIndex(namespace, startT, endT, &list)
					} else {
						appendDefaultNS = true
					}
				}
				if appendDefaultNS {
					namespace, ok := ns.Groups[m.defaultNamespace]
					if ok {
						m.findIndex(namespace, startT, endT, &list)
					}
				}
			}
		}
	}
	if len(list) == 0 {
		list = append(list, m.EmptyIndex())
	} else {
		sort.Strings(list)
	}
	return list
}

func (m *IndexManager) findIndex(namespace *indexGroup, start, end time.Time, list *[]string) {
	for _, entry := range namespace.List {
		if matchTimeRange(entry, start, end) {
			*list = append(*list, entry.Index)
		}
	}
	if namespace.Fixed != nil {
		*list = append(*list, namespace.Fixed.Index)
	}
	for _, key := range namespace.Groups {
		for _, entry := range key.List {
			if matchTimeRange(entry, start, end) {
				*list = append(*list, entry.Index)
			}
		}
		if key.Fixed != nil {
			*list = append(*list, key.Fixed.Index)
		}
	}
}

// EmptyIndex .
func (m *IndexManager) EmptyIndex() string {
	return fmt.Sprintf("%s-empty", m.cfg.IndexPrefix)
}

// IndexPrefix .
func (m *IndexManager) IndexPrefix() string { return m.cfg.IndexPrefix }

// IndexType .
func (m *IndexManager) IndexType() string { return m.cfg.IndexType }

// MetricNames .
func (m *IndexManager) MetricNames() (names []string) {
	v := m.indices.Load()
	if v != nil {
		indices := v.(map[string]*indexGroup)
		for index := range indices {
			names = append(names, index)
		}
		sort.Strings(names)
	}
	return names
}

// RequestTimeout .
func (m *IndexManager) RequestTimeout() time.Duration { return m.cfg.RequestTimeout }

// Client .
func (m *IndexManager) Client() *elastic.Client { return m.client }

// URLs .
func (m *IndexManager) URLs() string { return m.urls }

// EnableRollover .
func (m *IndexManager) EnableRollover() bool { return m.cfg.EnableRollover }
