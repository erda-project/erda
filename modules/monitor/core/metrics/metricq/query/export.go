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

package query

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/encoding/jsonx"
)

// Export 暂时没用到, 待移除
func (q *queryer) Export(ql, statement string, params map[string]interface{}, options url.Values, handle func(id string, data []byte) error) error {
	parser, start, end, _, err := q.buildTSQLParser(ql, statement, params, options)
	if err != nil {
		return err
	}
	sources, boolQuery, searchSource, err := parser.ParseRawQuery()
	metrics, clusters := getMetricsAndClustersFromSources(sources)
	indices := q.index.GetReadIndices(metrics, clusters, start, end)
	for _, c := range clusters {
		boolQuery.Filter(elastic.NewTermQuery(ClusterNameKey, c))
	}
	if len(indices) == 1 {
		if strings.HasSuffix(indices[0], "-empty") {
			boolQuery.Filter(elastic.NewTermQuery(TagKey+".not_exist", "_not_exist"))
		}
	}
	var sid string
	next := true
	var total int64
	for next {
		resp, err := q.esScrollRequest(indices, searchSource, sid)
		if err != nil {
			return err
		}
		sid, total, next, err = q.handleResp(resp, total, 10000, handle)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *queryer) esScrollRequest(indices []string, searchSource *elastic.SearchSource, sid string) (*elastic.SearchResult, error) {
	req := q.index.Client().Scroll(indices...).Scroll("1m").IgnoreUnavailable(true).AllowNoIndices(true)
	if len(sid) > 0 {
		req = req.ScrollId(sid)
	} else {
		req = req.SearchSource(searchSource)
	}
	context, cancel := context.WithTimeout(context.Background(), q.index.RequestTimeout())
	defer cancel()
	resp, err := req.Do(context)
	if err != nil || (resp != nil && resp.Error != nil) {
		if err == io.EOF || len(indices) <= 0 || (len(indices) == 1 && indices[0] == q.index.EmptyIndex()) {
			return nil, nil
		}
		if resp != nil && resp.Error != nil {
			return nil, fmt.Errorf("fail to request storage: %s", jsonx.MarshalAndIntend(resp.Error))
		}
		return nil, fmt.Errorf("fail to request storage: %s", err)
	}
	return resp, nil
}

func (q *queryer) handleResp(resp *elastic.SearchResult, total, limit int64, handle func(id string, data []byte) error) (string, int64, bool, error) {
	if resp != nil && resp.Hits != nil {
		hits := *resp.Hits
		for _, hit := range hits.Hits {
			if total >= limit {
				return "", total, false, nil
			}
			if hit.Source != nil {
				handle(hit.Uid, []byte(*hit.Source))
			} else {
				handle(hit.Uid, nil)
			}
			total++
		}
		if len(hits.Hits) > 0 {
			return resp.ScrollId, total, true, nil
		}
	}
	return "", total, false, nil
}
