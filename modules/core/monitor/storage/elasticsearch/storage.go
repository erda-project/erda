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

	"github.com/erda-project/erda/modules/core/monitor/storage"
	"github.com/olivere/elastic"
)

type esStorage struct {
	client       *elastic.Client
	typ          string
	queryTimeout string
	writeTimeout string
}

var _ storage.Storage = (*esStorage)(nil)

// Data .
type Data struct {
	Timestamp int64                  `json:"timestamp"`  // unix timestamp, nanosecond
	Date      int64                  `json:"@timestamp"` // unix timestamp, millisecond
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
}

const timestampKey = "timestamp"

func (s *esStorage) Mode() storage.Mode { return storage.ModeReadWrite }

func (s *esStorage) Query(ctx context.Context, sel *storage.Selector) (*storage.QueryResult, error) {
	searchSource := s.getSearchSource(sel)
	resp, err := s.client.Search(sel.PartitionKeys...).
		IgnoreUnavailable(true).AllowNoIndices(true).Size(storage.DefaultLimit).Sort(timestampKey, true).
		SearchSource(searchSource).Do(ctx)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Hits == nil {
		return &storage.QueryResult{}, nil
	}
	result := &storage.QueryResult{
		OriginalTotal: resp.Hits.TotalHits,
	}
	result.Values = parseHits(resp.Hits.Hits, sel.Matcher)
	return result, nil
}

func (s *esStorage) getSearchSource(sel *storage.Selector) *elastic.SearchSource {
	searchSource := elastic.NewSearchSource()
	query := elastic.NewBoolQuery().Filter(elastic.NewRangeQuery(timestampKey).Gte(sel.StartTime).Lt(sel.EndTime))
	for k, v := range sel.Labels {
		query = query.Filter(elastic.NewTermQuery(k, v))
	}
	return searchSource.Query(query)
}

func parseHits(hits []*elastic.SearchHit, matcher storage.Matcher) (list []*storage.Data) {
	for _, hit := range hits {
		if hit.Source == nil {
			continue
		}
		data, err := parseData(*hit.Source)
		if err != nil {
			continue
		}
		if matcher != nil && !matcher.Match(data) {
			continue
		}
		list = append(list, data)
	}
	return list
}

func parseData(byts []byte) (*storage.Data, error) {
	var data Data
	err := json.Unmarshal(byts, &data)
	if err != nil {
		return nil, err
	}
	return &storage.Data{
		Timestamp: data.Timestamp,
		Labels:    data.Tags,
		Fields:    data.Fields,
	}, nil
}
