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
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/olivere/elastic"
)

// WaitIndicesLoad .
func (p *provider) WaitAndGetIndices(ctx context.Context) map[string]*IndexGroup {
	for {
		v := p.indices.Load()
		if v == nil {
			// Wait for the index to complete loading
			select {
			case <-time.After(1 * time.Second):
			case <-ctx.Done():
				return nil
			}
			continue
		}
		return v.(map[string]*IndexGroup)
	}
}

func (p *provider) AllIndices() map[string]*IndexGroup {
	v, _ := p.indices.Load().(map[string]*IndexGroup)
	return v
}

func (p *provider) ReloadIndices() error {
	ch := make(chan error)
	p.reloadCh <- ch
	return <-ch
}

func (p *provider) WatchLoadEvent(f func(map[string]*IndexGroup)) {
	p.listeners = append(p.listeners, f)
}

// GetReadIndices .
func (p *provider) GetReadIndices(metrics []string, namespaces []string, start, end int64) (list []string) {
	if len(metrics) <= 0 {
		return []string{p.EmptyIndex()}
	}
	v := p.indices.Load()
	if v == nil {
		for _, name := range metrics {
			list = append(list, p.Cfg.IndexPrefix+"-"+normalizeIndexSegmentName(strings.ToLower(name))+"-*")
		}
	} else {
		indices := v.(map[string]*IndexGroup)
		startT := time.Unix(start/1000, (start%1000)*int64(time.Millisecond))
		endT := time.Unix(end/1000, (end%1000)*int64(time.Millisecond)+int64(time.Millisecond)-1)
		for _, name := range metrics {
			name = normalizeIndexSegmentName(strings.ToLower(name))
			ns, ok := indices[name]
			if !ok {
				continue
			}
			if len(namespaces) == 0 {
				for _, namespace := range ns.Groups {
					p.findIndex(namespace, startT, endT, &list)
				}
			} else {
				var appendDefaultNS bool
				for _, n := range namespaces {
					n = normalizeIndexSegmentName(n)
					if n == p.Cfg.DefaultNamespace {
						appendDefaultNS = true
						continue
					}
					namespace, ok := ns.Groups[n]
					if ok {
						p.findIndex(namespace, startT, endT, &list)
					} else {
						appendDefaultNS = true
					}
				}
				if appendDefaultNS {
					namespace, ok := ns.Groups[p.Cfg.DefaultNamespace]
					if ok {
						p.findIndex(namespace, startT, endT, &list)
					}
				}
			}
		}
	}
	if len(list) == 0 {
		list = append(list, p.EmptyIndex())
	} else {
		sort.Strings(list)
	}
	return list
}

func (p *provider) findIndex(namespace *IndexGroup, start, end time.Time, list *[]string) {
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

func matchTimeRange(entry *IndexEntry, start, end time.Time) bool {
	if (entry.MinT.IsZero() || entry.MinT.Before(end) || entry.MinT.Equal(end)) &&
		(entry.MaxT.IsZero() || entry.MaxT.After(start) || entry.MaxT.Equal(start)) {
		return true
	}
	return false
}

func normalizeIndexSegmentName(s string) string { return strings.Replace(s, "-", "_", -1) }

func (p *provider) MetricNames() (names []string) {
	v := p.indices.Load()
	if v != nil {
		indices := v.(map[string]*IndexGroup)
		for index := range indices {
			names = append(names, index)
		}
		sort.Strings(names)
	}
	return names
}

func (p *provider) EmptyIndex() string            { return fmt.Sprintf("%s-empty", p.Cfg.IndexPrefix) }
func (p *provider) IndexPrefix() string           { return p.Cfg.IndexPrefix }
func (p *provider) RequestTimeout() time.Duration { return p.Cfg.RequestTimeout }
func (p *provider) QueryIndexTimeRange() bool     { return p.Cfg.QueryIndexTimeRange }
func (p *provider) Client() *elastic.Client       { return p.ES.Client() }
func (p *provider) URLs() string                  { return p.ES.URL() }
