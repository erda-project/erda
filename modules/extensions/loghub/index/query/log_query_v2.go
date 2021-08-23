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

package query

import (
	"encoding/json"
	"time"

	"github.com/olivere/elastic"

	logs "github.com/erda-project/erda/modules/core/monitor/log"
)

func (c *ESClient) getBoolQueryV2(req *LogRequest) *elastic.BoolQuery {
	boolQuery := c.getTagsBoolQuery(req)
	start := req.Start * int64(time.Millisecond)
	end := req.End * int64(time.Millisecond)
	boolQuery = boolQuery.Filter(elastic.NewRangeQuery("timestamp").Gte(start).Lte(end))
	if len(req.Query) > 0 {
		byts, _ := json.Marshal(req.Query)
		boolQuery = boolQuery.Filter(elastic.NewQueryStringQuery("content:" + string(byts)))
	}
	return boolQuery
}

func (c *ESClient) searchLogsV2(req *LogSearchRequest, timeout time.Duration) (*LogQueryResponse, error) {
	boolQuery := c.getBoolQueryV2(&req.LogRequest)
	searchSource := c.getSearchSource(req, boolQuery)
	if len(req.Sort) <= 0 {
		searchSource.Sort("timestamp", true).Sort("offset", true)
	}
	if req.Debug {
		c.printSearchSource(searchSource)
	}
	total, hits, err := c.doSearchLogs(req, searchSource, timeout)
	if err != nil {
		return nil, err
	}
	resp := &LogQueryResponse{
		Total: total,
	}
	for _, hit := range hits {
		if hit.Source == nil {
			continue
		}
		var log logs.Log
		err := json.Unmarshal([]byte(*hit.Source), &log)
		if err != nil {
			continue
		}
		c.setModule(&log)
		log.Timestamp = log.Timestamp / int64(time.Millisecond)
		resp.Data = append(resp.Data, &log)
	}
	return resp, nil
}

func (c *ESClient) statisticLogsV2(req *LogStatisticRequest, timeout time.Duration, name string) (*LogStatisticResponse, error) {
	boolQuery := c.getBoolQueryV2(&req.LogRequest)
	searchSource := elastic.NewSearchSource().Query(boolQuery)
	searchSource.Size(0)
	interval := req.Interval
	if req.Points > 0 {
		interval = (req.End - req.Start) / req.Points
	}
	intervalMillisecond := interval
	start := req.Start * int64(time.Millisecond)
	end := req.End * int64(time.Millisecond)
	interval = interval * int64(time.Millisecond)
	searchSource = searchSource.Aggregation("timestamp",
		elastic.NewHistogramAggregation().
			Field("timestamp").
			Interval(float64(interval)).
			MinDocCount(0).
			Offset(float64(start%interval)).
			ExtendedBounds(float64(start), float64(end)),
	)
	if req.Debug {
		c.printSearchSource(searchSource)
	}
	resp, err := c.doRequest(searchSource, timeout)
	if err != nil {
		return nil, err
	}
	result := newLogStatisticResponse(intervalMillisecond, resp.TotalHits(), name)
	if resp.Aggregations == nil {
		return result, nil
	}
	histogram, ok := resp.Aggregations.Histogram("timestamp")
	if !ok {
		return result, nil
	}
	list := result.Results[0].Data[0].Count.Data
	for i, b := range histogram.Buckets {
		if req.Points > 0 && int64(i+1) > req.Points && len(list) > 0 {
			last := len(list) - 1
			list[last] = list[last] + float64(b.DocCount)
			continue
		}
		result.Time = append(result.Time, int64(b.Key)/int64(time.Millisecond))
		list = append(list, float64(b.DocCount))
	}
	result.Results[0].Data[0].Count.Data = list
	return result, nil
}
