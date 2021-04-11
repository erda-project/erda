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

package admin_tools

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/lang/size"
)

func (p *provider) showIndicesByDate(w http.ResponseWriter, param struct {
	Wildcard string `param:"wildcard"`
	Format   string `query:"format"`
}) interface{} {
	indices, err := p.getIndices(param.Wildcard)
	if err != nil {
		return api.Errors.Internal(err)
	}
	list := p.sortIndices(indices)
	if param.Format == "json" {
		return api.Success(list)
	}
	for _, item := range list {
		tm := time.Unix(item.Timestamp/1000, 0)
		line := fmt.Sprint("----------- ", tm.Format("2006-01-02T15:04:05Z07:00"), " ----------- ", len(item.Indices), "\n")
		w.Write([]byte(line))
		for _, index := range item.Indices {
			line := fmt.Sprintf("%s\t%d\t%s\n", index.Index, index.DocsCount, index.StoreSize)
			w.Write([]byte(line))
		}
		w.Write([]byte("---------------------------------------------------\n\n"))
	}
	return nil
}

// IndexEntry .
type IndexEntry struct {
	Timestamp      int64 `json:"timestamp"`
	StoreSizeValue int64 `json:"store_size_value"`

	Health       string `json:"health"`              // "green", "yellow", or "red"
	Status       string `json:"status"`              // "open" or "closed"
	Index        string `json:"index"`               // index name
	UUID         string `json:"uuid"`                // index uuid
	Pri          int    `json:"pri,string"`          // number of primary shards
	Rep          int    `json:"rep,string"`          // number of replica shards
	DocsCount    int    `json:"docs.count,string"`   // number of available documents
	DocsDeleted  int    `json:"docs.deleted,string"` // number of deleted documents
	StoreSize    string `json:"store.size"`          // store size of primaries & replicas, e.g. "4.6kb"
	PriStoreSize string `json:"pri.store.size"`      // store size of primaries, e.g. "230b"
}

// IndexEntries .
type IndexEntries []*IndexEntry

func (l IndexEntries) Len() int      { return len(l) }
func (l IndexEntries) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l IndexEntries) Less(i, j int) bool {
	a, b := l[i], l[j]
	return a.StoreSizeValue >= b.StoreSizeValue
}

func (p *provider) getIndices(wildcard string) ([]*IndexEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resps, err := p.es.CatIndices().Index(wildcard).Do(ctx)
	if err != nil {
		return nil, err
	}
	var list []*IndexEntry
	for _, item := range resps {
		idx := strings.LastIndex(item.Index, "-")
		if idx > 0 && idx+1 < len(item.Index) {
			timestamp, err := strconv.ParseInt(item.Index[idx+1:], 10, 64)
			if err == nil {
				list = append(list, newIndexEntry(timestamp, &item))
				continue
			}
		}
		list = append(list, newIndexEntry(0, &item))
	}
	return list, nil
}

func newIndexEntry(timestamp int64, cat *elastic.CatIndicesResponseRow) *IndexEntry {
	storeSize, _ := size.ParseBytes(cat.StoreSize)
	return &IndexEntry{
		Timestamp:      timestamp,
		StoreSizeValue: storeSize,
		Health:         cat.Health,
		Status:         cat.Status,
		Index:          cat.Index,
		UUID:           cat.UUID,
		Pri:            cat.Pri,
		Rep:            cat.Rep,
		DocsCount:      cat.DocsCount,
		DocsDeleted:    cat.DocsDeleted,
		StoreSize:      cat.StoreSize,
		PriStoreSize:   cat.PriStoreSize,
	}
}

// DateIndices .
type DateIndices struct {
	Timestamp int64         `json:"timestamp"`
	Indices   []*IndexEntry `json:"indices"`
}

func (p *provider) sortIndices(indices []*IndexEntry) []*DateIndices {
	dates := make(map[int64][]*IndexEntry)
	for _, index := range indices {
		t := index.Timestamp - index.Timestamp%(24*60*60*1000)
		dates[t] = append(dates[t], index)
	}
	var keys []int64
	for t := range dates {
		keys = append(keys, t)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] >= keys[j]
	})
	var list []*DateIndices
	for _, t := range keys {
		indices := dates[t]
		sort.Sort(IndexEntries(indices))
		list = append(list, &DateIndices{
			Timestamp: t,
			Indices:   indices,
		})
	}
	return list
}

func (p *provider) deleteIndices(param struct {
	Wildcard string `param:"wildcard"`
}) interface{} {
	if len(param.Wildcard) <= 0 {
		return api.Success(0)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	indices := strings.Split(param.Wildcard, ",")
	resp, err := p.es.DeleteIndex(indices...).Do(ctx)
	if err != nil {
		if e, ok := err.(*elastic.Error); ok {
			if e.Status == 404 {
				return api.Success(0)
			}
		}
		return api.Errors.Internal(err)
	}
	if !resp.Acknowledged {
		return api.Errors.Internal(fmt.Errorf("delete indices Acknowledged=false"))
	}
	p.L.Infof("delete indices %d, %v", len(indices), indices)
	return api.Success(len(indices))
}
