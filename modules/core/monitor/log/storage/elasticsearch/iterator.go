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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/encoding/jsonx"

	"github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/modules/core/monitor/log"
	"github.com/erda-project/erda/modules/core/monitor/log/storage"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
)

const useScrollQuery = false
const useInMemContentFilter = true

func (p *provider) Iterator(ctx context.Context, sel *storage.Selector) (storekit.Iterator, error) {
	// TODO check org
	indices := p.Loader.Indices(ctx, sel.Start, sel.End, loader.KeyPath{
		Recursive: true,
	})
	var matcher func(data *pb.LogItem) bool
	if useInMemContentFilter {
		for _, filter := range sel.Filters {
			val, _ := filter.Value.(string)
			if filter.Key != "content" || len(val) <= 0 {
				continue
			}
			switch filter.Op {
			case storage.EQ:
				matcher = func(data *pb.LogItem) bool {
					return data.Content == val
				}
			case storage.REGEXP:
				regex, err := regexp.Compile(val)
				if err != nil {
					p.Log.Debugf("invalid regexp %q", val)
					return storekit.EmptyIterator{}, nil
				}
				matcher = func(data *pb.LogItem) bool {
					return regex.MatchString(data.Content)
				}
			}
		}
	}

	if useScrollQuery {
		searchSource := getSearchSource(sel.Start, sel.End, sel)
		if sel.Debug {
			source, _ := searchSource.Source()
			fmt.Printf("indices: %v\nsearchSource: %s\n", strings.Join(indices, ","), jsonx.MarshalAndIndent(source))
		}
		return elasticsearch.NewScrollIterator(
			ctx, p.client, p.Cfg.QueryTimeout,
			p.Cfg.ReadPageSize, indices, []*elasticsearch.SortItem{
				{
					Key:       "timestamp",
					Ascending: true,
				},
				{
					Key:       "offset",
					Ascending: true,
				},
			},
			func() (*elastic.SearchSource, error) { return searchSource, nil },
			decodeFunc(sel.Start, sel.End, matcher),
		)
	}
	if sel.Debug {
		searchSource := getSearchSource(sel.Start, sel.End, sel)
		source, _ := searchSource.Source()
		fmt.Printf("indices: %v\nsearchSource: %s\n", strings.Join(indices, ","), jsonx.MarshalAndIndent(source))
	}
	return elasticsearch.NewSearchIterator(
		ctx, p.client, p.Cfg.QueryTimeout,
		p.Cfg.ReadPageSize, indices, []*elasticsearch.SortItem{
			{
				Key:       "timestamp",
				Ascending: true,
			},
			{
				Key:       "offset",
				Ascending: true,
			},
		},
		func() (*elastic.SearchSource, error) {
			return getSearchSource(sel.Start, sel.End, sel), nil
		},
		decodeFunc(sel.Start, sel.End, matcher),
	)
}

func getSearchSource(start, end int64, sel *storage.Selector) *elastic.SearchSource {
	searchSource := elastic.NewSearchSource()
	query := elastic.NewBoolQuery().Filter(elastic.NewRangeQuery("timestamp").Gte(start).Lt(end))
	for _, filter := range sel.Filters {
		val, ok := filter.Value.(string)
		if !ok {
			continue
		}
		if useInMemContentFilter {
			if filter.Key == "content" {
				continue
			}
		}
		switch filter.Op {
		case storage.EQ:
			query = query.Filter(elastic.NewTermQuery(filter.Key, val))
		case storage.REGEXP:
			query = query.Filter(elastic.NewRegexpQuery(filter.Key, val))
		}
	}
	return searchSource.Query(query)
}

var skip = errors.New("skip")

func decodeFunc(start, end int64, matcher func(data *pb.LogItem) bool) func(body []byte) (interface{}, error) {
	return func(body []byte) (interface{}, error) {
		var data log.Log
		err := json.Unmarshal(body, &data)
		if err != nil {
			return nil, err
		}
		if data.Timestamp < start || data.Timestamp >= end {
			return nil, skip
		}
		item := &pb.LogItem{
			Source:    data.Source,
			Id:        data.ID,
			Stream:    data.Stream,
			Timestamp: strconv.FormatInt(data.Timestamp, 10),
			UnixNano:  data.Timestamp,
			Offset:    data.Offset,
			Content:   data.Content,
			Level:     data.Tags["level"],
			RequestId: data.Tags["request_id"],
		}
		if matcher != nil && !matcher(item) {
			return nil, skip
		}
		return item, nil
	}
}
