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
	"fmt"
	"strings"
	"time"

	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/encoding/jsonx"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric/query"
	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
	indexloader "github.com/erda-project/erda/internal/tools/monitor/core/storekit/elasticsearch/index/loader"
)

// Interface .
type Interface interface {
	SearchRaw(indices []string, searchSource *elastic.SearchSource) (*elastic.SearchResult, error)
	QueryRaw(metrics, clusters []string, start, end int64, searchSource *elastic.SearchSource) (*elastic.SearchResult, error)
}

func (p *provider) Query(ctx context.Context, q tsql.Query) (*model.ResultSet, error) {
	var err error
	metrics, clusters := getMetricsAndClustersFromSources(q.Sources())
	start, end := q.Timestamp()

	indices := p.getIndices(metrics, start, end)
	for _, c := range clusters {
		q.AppendBoolFilter(model.TagKey+".cluster_name", c)
	}
	if len(indices) == 1 {
		if strings.HasSuffix(indices[0], "-empty") {
			q.AppendBoolFilter(model.TagKey+".not_exist", "_not_exist")
		}
	}
	searchSource, ok := q.SearchSource().(*elastic.SearchSource)
	if !ok {
		return nil, fmt.Errorf("invalid search source")
	}

	result := &model.ResultSet{}

	if q.Debug() {
		var source interface{}
		if searchSource != nil {
			source, err = searchSource.Source()
			if err != nil {
				return nil, fmt.Errorf("invalid search source: %s", err)
			}
		}

		result.Details = query.ElasticSearchCURL(p.Loader.URLs(), indices, source)
		fmt.Println(result.Details)
		return result, nil
	}

	var resp *elastic.SearchResult
	if searchSource != nil {
		now := time.Now()
		resp, err = p.search(ctx, indices, searchSource)
		if err != nil {
			return nil, err
		}
		result.Elapsed.Search = time.Now().Sub(now)
	}

	result.Data, err = q.ParseResult(ctx, resp)
	if err != nil {
		return nil, err
	}
	return result, nil

}

func (p *provider) SearchRaw(indices []string, searchSource *elastic.SearchSource) (*elastic.SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.Loader.RequestTimeout())
	defer cancel()
	return p.Loader.Client().Search(indices...).
		IgnoreUnavailable(true).AllowNoIndices(true).
		SearchSource(searchSource).Do(ctx)
}

func (p *provider) QueryRaw(metrics, clusters []string, start, end int64, searchSource *elastic.SearchSource) (*elastic.SearchResult, error) {
	indices := p.getIndices(metrics, start, end)
	return p.SearchRaw(indices, searchSource)
}

func (p *provider) search(ctx context.Context, indices []string, searchSource *elastic.SearchSource) (*elastic.SearchResult, error) {
	context, cancel := context.WithTimeout(ctx, p.Loader.RequestTimeout())
	defer cancel()

	resp, err := execution(context, p.Loader.Client(), indices, searchSource)
	if err != nil || (resp != nil && resp.Error != nil) {
		if len(indices) <= 0 || (len(indices) == 1 && strings.HasSuffix(indices[0], "-empty")) {
			return nil, nil
		}
		if resp != nil && resp.Error != nil {
			return nil, fmt.Errorf("fail to request storage: %s", jsonx.MarshalAndIndent(resp.Error))
		}
		return nil, fmt.Errorf("fail to request storage: %s", err)
	}
	return resp, nil
}

var execution = func(ctx context.Context, client *elastic.Client, indices []string, searchSource *elastic.SearchSource) (*elastic.SearchResult, error) {
	return client.Search(indices...).
		IgnoreUnavailable(true).AllowNoIndices(true).
		SearchSource(searchSource).Do(ctx)
}

func (p *provider) getIndices(metrics []string, start, end int64) []string {
	keys := make([]indexloader.KeyPath, len(metrics)+1)
	for i, item := range metrics {
		keys[i] = indexloader.KeyPath{
			Keys:      []string{item},
			Recursive: true,
		}
	}
	keys[len(metrics)] = indexloader.KeyPath{}
	start = start * int64(time.Millisecond)
	end = end*int64(time.Millisecond) + (int64(time.Millisecond) - 1)
	return p.Loader.Indices(context.Background(), start, end, keys...)
}

func getMetricsAndClustersFromSources(sources []*model.Source) (metrics []string, clusters []string) {
	for _, source := range sources {
		if len(source.Name) > 0 {
			metrics = append(metrics, source.Name)
		}
		if len(source.Database) > 0 {
			clusters = append(clusters, source.Database)
		}
	}
	return metrics, clusters
}

func (p *provider) Select(metric []string) bool {
	return true
}

func (p *provider) QueryExternal(ctx context.Context, q tsql.Query) (*model.ResultSet, error) {
	return nil, fmt.Errorf("not support")
}
