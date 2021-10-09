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
	"io"
	"strconv"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/modules/core/monitor/storage"
	"github.com/olivere/elastic"
)

func (s *esStorage) Iterator(ctx context.Context, sel *storage.Selector) storage.Iterator {
	return &scrollIterator{
		ctx:          ctx,
		sel:          sel,
		searchSource: s.getSearchSource(sel),
		client:       s.client,
		timeout:      3 * time.Minute,
		pageSize:     4 * 1024,
	}
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

	scrollIDs    map[string]struct{}
	lastScrollID string
	dir          iteratorDir
	buffer       []*storage.Data
	value        *storage.Data
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
		it.err = storage.ErrOpNotSupported
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
		it.err = storage.ErrOpNotSupported
		return false
	}
	if it.yield() {
		return true
	}
	it.fetch(iteratorBackward)
	return it.yield()
}

func (it *scrollIterator) Value() *storage.Data { return it.value }
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
				resp, it.err = it.client.Scroll(it.sel.PartitionKeys...).KeepAlive(keepalive).
					IgnoreUnavailable(true).AllowNoIndices(true).
					SearchSource(it.searchSource).Size(it.pageSize).Sort(timestampKey, ascending).Do(ctx)
				if it.err != nil {
					return it.err
				}
			} else {
				resp, it.err = it.client.Scroll(it.sel.PartitionKeys...).ScrollId(it.lastScrollID).KeepAlive(keepalive).
					IgnoreUnavailable(true).AllowNoIndices(true).
					Size(it.pageSize).Do(ctx)
				if it.err != nil {
					return it.err
				}
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
			it.buffer = parseHits(resp.Hits.Hits, it.sel.Matcher)
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
			it.err = storage.ErrIteratorClosed
		}
		return true
	}
	return false
}
