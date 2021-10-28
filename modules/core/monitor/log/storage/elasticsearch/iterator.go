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

package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/encoding/jsonx"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/modules/core/monitor/log"
	"github.com/erda-project/erda/modules/core/monitor/log/storage"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
)

func getSearchSource(start, end int64, includeEnd bool, sel *storage.Selector, condition func(query *elastic.BoolQuery) *elastic.BoolQuery) *elastic.SearchSource {
	searchSource := elastic.NewSearchSource()
	query := elastic.NewBoolQuery()
	if includeEnd {
		query = query.Filter(elastic.NewRangeQuery("timestamp").Gte(start).Lte(end))
	} else {
		query = query.Filter(elastic.NewRangeQuery("timestamp").Gte(start).Lt(end))
	}
	for _, filter := range sel.Filters {
		val, ok := filter.Value.(string)
		if !ok {
			continue
		}
		switch filter.Op {
		case storage.EQ:
			query = query.Filter(elastic.NewTermQuery(filter.Key, val))
		case storage.REGEXP:
			query = query.Filter(elastic.NewRegexpQuery(filter.Key, val))
		}
	}
	if condition != nil {
		query = condition(query)
	}
	return searchSource.Query(query)
}

const useScrollQuery = false

func (p *provider) Iterator(ctx context.Context, sel *storage.Selector) (storekit.Iterator, error) {
	// TODO check org
	indices := p.Loader.Indices(ctx, sel.Start, sel.End, loader.KeyPath{
		Recursive: true,
	})
	if useScrollQuery {
		searchSource := getSearchSource(sel.Start, sel.End, false, sel, nil)
		if sel.Debug {
			source, _ := searchSource.Source()
			fmt.Printf("indices: %v\nsearchSource: %s\n", strings.Join(indices, ","), jsonx.MarshalAndIndent(source))
		}
		return &scrollIterator{
			ctx:          ctx,
			sel:          sel,
			searchSource: searchSource,
			client:       p.client,
			timeout:      p.Cfg.QueryTimeout,
			pageSize:     p.Cfg.ReadPageSize,
			indices:      indices,
		}, nil
	}
	return &searchIterator{
		ctx:        ctx,
		sel:        sel,
		client:     p.client,
		timeout:    p.Cfg.QueryTimeout,
		pageSize:   p.Cfg.ReadPageSize,
		indices:    indices,
		lastStart:  sel.Start,
		lastEnd:    sel.End,
		lastOffset: 0,
	}, nil
}

type iteratorDir int8

const (
	iteratorInitial = iota
	iteratorForward
	iteratorBackward
)

type scrollIterator struct {
	log          logs.Logger
	ctx          context.Context
	sel          *storage.Selector
	searchSource *elastic.SearchSource
	client       *elastic.Client
	timeout      time.Duration
	pageSize     int
	indices      []string

	scrollIDs    map[string]struct{}
	lastScrollID string
	dir          iteratorDir
	buffer       []*pb.LogItem
	value        *pb.LogItem
	err          error
	closed       bool
}

func (it *scrollIterator) First() bool {
	if it.checkClosed() {
		return false
	}
	it.release()
	it.fetch(iteratorForward)
	return it.yield()
}

func (it *scrollIterator) Last() bool {
	if it.checkClosed() {
		return false
	}
	it.release()
	it.fetch(iteratorBackward)
	return it.yield()
}

func (it *scrollIterator) Next() bool {
	if it.checkClosed() {
		return false
	}
	if it.dir == iteratorBackward {
		it.err = storekit.ErrOpNotSupported
		return false
	}
	if it.yield() {
		return true
	}
	it.fetch(iteratorForward)
	return it.yield()
}

func (it *scrollIterator) Prev() bool {
	if it.checkClosed() {
		return false
	}
	if it.dir == iteratorForward {
		it.err = storekit.ErrOpNotSupported
		return false
	}
	if it.yield() {
		return true
	}
	it.fetch(iteratorBackward)
	return it.yield()
}

func (it *scrollIterator) Value() storekit.Data { return it.value }
func (it *scrollIterator) Error() error {
	if it.err == io.EOF {
		return nil
	}
	return it.err
}

func (it *scrollIterator) release() (err error) {
	var list []string
	for id := range it.scrollIDs {
		if len(id) > 0 {
			list = append(list, id)
		}
	}
	if len(list) > 0 {
		_, err = it.client.ClearScroll(list...).Do(context.TODO())
		if err != nil {
			it.log.Errorf("failed to clear scroll: %s", err)
		}
	}
	it.scrollIDs, it.lastScrollID = nil, ""
	it.buffer = nil
	it.value = nil
	return nil
}

func (it *scrollIterator) fetch(dir iteratorDir) error {
	if len(it.indices) <= 0 {
		it.err = io.EOF
		return it.err
	}
	minutes := int64(it.timeout.Minutes())
	if minutes < 1 {
		minutes = 1
	}
	keepalive := strconv.FormatInt(minutes, 10) + "m"

	it.dir = dir
	it.buffer = nil
	for it.err == nil && len(it.buffer) <= 0 {
		func() error {
			// do query
			ctx, cancel := context.WithTimeout(it.ctx, it.timeout)
			defer cancel()
			var resp *elastic.SearchResult
			if len(it.lastScrollID) <= 0 {
				var ascending bool
				if it.dir != iteratorBackward {
					ascending = true
				}
				resp, it.err = it.client.Scroll(it.indices...).KeepAlive(keepalive).
					IgnoreUnavailable(true).AllowNoIndices(true).
					SearchSource(it.searchSource).Size(it.pageSize).Sort("timestamp", ascending).Sort("offset", ascending).Do(ctx)
			} else {
				resp, it.err = it.client.Scroll(it.indices...).ScrollId(it.lastScrollID).KeepAlive(keepalive).
					IgnoreUnavailable(true).AllowNoIndices(true).
					Size(it.pageSize).Do(ctx)
			}
			if it.err != nil {
				return it.err
			}

			// save scrollID
			if it.scrollIDs == nil {
				it.scrollIDs = make(map[string]struct{})
			}
			if resp != nil {
				it.scrollIDs[resp.ScrollId] = struct{}{}
				it.lastScrollID = resp.ScrollId
			}
			if resp == nil || resp.Hits == nil || len(resp.Hits.Hits) <= 0 {
				it.err = io.EOF
				return it.err
			}

			// parse result
			it.buffer = parseHits(resp.Hits.Hits, it.sel.Start, it.sel.End)
			return nil
		}()
	}
	return nil
}

func (it *scrollIterator) yield() bool {
	if len(it.buffer) > 0 {
		it.value = it.buffer[0]
		it.buffer = it.buffer[1:]
		return true
	}
	return false
}

func (it *scrollIterator) Close() error {
	it.closed = true
	it.release()
	return nil
}

func (it *scrollIterator) checkClosed() bool {
	if it.closed {
		if it.err == nil {
			it.err = storekit.ErrIteratorClosed
		}
		return true
	}
	select {
	case <-it.ctx.Done():
		if it.err == nil {
			it.err = storekit.ErrIteratorClosed
		}
		return true
	default:
	}
	return false
}

type searchIterator struct {
	log logs.Logger
	ctx context.Context
	sel *storage.Selector

	lastStart  int64
	lastEnd    int64
	lastOffset int64

	client   *elastic.Client
	timeout  time.Duration
	pageSize int
	indices  []string

	dir    iteratorDir
	buffer []*pb.LogItem
	value  *pb.LogItem
	err    error
	closed bool
}

func (it *searchIterator) First() bool {
	if it.checkClosed() {
		return false
	}
	it.lastStart = it.sel.Start
	it.lastEnd = it.sel.End
	it.lastOffset = 0
	it.fetch(iteratorForward)
	return it.yield()
}

func (it *searchIterator) Last() bool {
	if it.checkClosed() {
		return false
	}
	it.lastStart = it.sel.Start
	it.lastEnd = it.sel.End
	it.lastOffset = 0
	it.fetch(iteratorBackward)
	return it.yield()
}

func (it *searchIterator) Next() bool {
	if it.checkClosed() {
		return false
	}
	if it.dir == iteratorBackward {
		it.err = storekit.ErrOpNotSupported
		return false
	}
	if it.yield() {
		return true
	}
	it.fetch(iteratorForward)
	return it.yield()
}

func (it *searchIterator) Prev() bool {
	if it.checkClosed() {
		return false
	}
	if it.dir == iteratorForward {
		it.err = storekit.ErrOpNotSupported
		return false
	}
	if it.yield() {
		return true
	}
	it.fetch(iteratorBackward)
	return it.yield()
}

func (it *searchIterator) Value() storekit.Data { return it.value }
func (it *searchIterator) Error() error {
	if it.err == io.EOF {
		return nil
	}
	return it.err
}

func (it *searchIterator) fetch(dir iteratorDir) error {
	if len(it.indices) <= 0 {
		it.err = io.EOF
		return it.err
	}
	ms := int64(it.timeout.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	timeout := strconv.FormatInt(ms, 10) + "ms"

	var ascending bool
	if dir != iteratorBackward {
		ascending = true
	}

	it.dir = dir
	it.buffer = nil
	for it.err == nil && len(it.buffer) <= 0 && it.lastStart < it.lastEnd {
		func() error {
			// do query
			ctx, cancel := context.WithTimeout(it.ctx, it.timeout)
			defer cancel()
			var resp *elastic.SearchResult

			var searchSource *elastic.SearchSource
			if ascending {
				searchSource = getSearchSource(it.lastStart, it.lastEnd, false, it.sel, func(query *elastic.BoolQuery) *elastic.BoolQuery {
					if it.lastOffset > 0 {
						query.Filter(elastic.NewRangeQuery("offset").Gt(it.lastOffset))
					}
					return query
				})
			} else {
				searchSource = getSearchSource(it.lastStart, it.lastEnd, it.lastEnd != it.sel.End, it.sel, func(query *elastic.BoolQuery) *elastic.BoolQuery {
					if it.lastOffset > 0 {
						query.Filter(elastic.NewRangeQuery("offset").Lt(it.lastOffset))
					}
					return query
				})
			}
			resp, it.err = it.client.Search(it.indices...).IgnoreUnavailable(true).AllowNoIndices(true).Timeout(timeout).
				SearchSource(searchSource).Size(it.pageSize).Sort("timestamp", ascending).Sort("offset", ascending).Do(ctx)
			if it.err != nil {
				return it.err
			}
			if resp == nil || resp.Hits == nil || len(resp.Hits.Hits) <= 0 {
				it.err = io.EOF
				return it.err
			}
			it.buffer = parseHits(resp.Hits.Hits, it.sel.Start, it.sel.End)
			num := len(it.buffer)
			if num > 0 {
				last := it.buffer[num-1]
				if ascending {
					it.lastStart = last.Timestamp
					// maybe offset is zero forever, so increase lastStart to avoid query duplicated log
					if it.lastOffset == last.Offset {
						it.lastStart++
					}

				} else {
					it.lastEnd = last.Timestamp
					// maybe offset is zero forever, so decrease lastEnd to avoid query duplicated log
					if it.lastOffset == last.Offset {
						it.lastEnd--
					}
				}
				it.lastOffset = last.Offset
			} else {
				it.err = io.EOF
				return it.err
			}
			return nil
		}()
	}
	return nil
}

func (it *searchIterator) yield() bool {
	if len(it.buffer) > 0 {
		it.value = it.buffer[0]
		it.buffer = it.buffer[1:]
		return true
	}
	return false
}

func (it *searchIterator) Close() error {
	it.closed = true
	return nil
}

func (it *searchIterator) checkClosed() bool {
	if it.closed {
		if it.err == nil {
			it.err = storekit.ErrIteratorClosed
		}
		return true
	}
	select {
	case <-it.ctx.Done():
		if it.err == nil {
			it.err = storekit.ErrIteratorClosed
		}
		return true
	default:
	}
	return false
}

func parseHits(hits []*elastic.SearchHit, start, end int64) (list []*pb.LogItem) {
	for _, hit := range hits {
		if hit.Source == nil {
			continue
		}
		data, err := parseData(*hit.Source)
		if err != nil {
			continue
		}
		if start <= data.Timestamp && data.Timestamp < end {
			list = append(list, data)
		}
	}
	return list
}

func parseData(byts []byte) (*pb.LogItem, error) {
	var data log.Log
	err := json.Unmarshal(byts, &data)
	if err != nil {
		return nil, err
	}
	return &pb.LogItem{
		Source:    data.Source,
		Id:        data.ID,
		Stream:    data.Stream,
		Timestamp: data.Timestamp,
		Offset:    data.Offset,
		Content:   data.Content,
		Level:     data.Tags["level"],
		RequestId: data.Tags["request_id"],
	}, nil
}
