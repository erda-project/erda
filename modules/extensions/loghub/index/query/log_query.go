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
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/encoding/jsonx"

	"github.com/erda-project/erda-infra/providers/i18n"
	logs "github.com/erda-project/erda/modules/core/monitor/log"
)

// Tag .
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// LogField .
type LogField struct {
	FieldName          string `json:"fieldName"`
	SupportAggregation bool   `json:"supportAggregation"`
	Display            bool   `json:"display"`
}

// LogRequest .
type LogRequest struct {
	OrgID       int64
	ClusterName string
	Addon       string
	Start       int64
	End         int64
	Filters     []*Tag
	Query       string
	Debug       bool
	Lang        i18n.LanguageCodes
}

// LogSearchRequest .
type LogSearchRequest struct {
	LogRequest
	Page      int64
	Size      int64
	Sort      string
	Highlight bool
}

// LogStatisticRequest .
type LogStatisticRequest struct {
	LogRequest
	Interval int64
	Points   int64
}

type LogItem struct {
	Source    *logs.Log           `json:"source"`
	Highlight map[string][]string `json:"highlight"`
}

// LogQueryResponse .
type LogQueryResponse struct {
	Expends map[string]interface{} `json:"expends"`
	Total   int64                  `json:"total"`
	Data    []*LogItem             `json:"data"`
}

// LogStatisticResponse .
type LogStatisticResponse struct {
	Expends  map[string]interface{} `json:"expends"`
	Title    string                 `json:"title"`
	Total    int64                  `json:"total"`
	Interval int64                  `json:"interval"`
	Time     []int64                `json:"time"`
	Results  []*LogStatisticResult  `json:"results"`
}

func newLogStatisticResponse(interval, total int64, name string) *LogStatisticResponse {
	return &LogStatisticResponse{
		Title:    name,
		Total:    total,
		Interval: interval,
		Results: []*LogStatisticResult{
			{
				Name: "count",
				Data: []*CountHistogram{
					{
						Count: ArrayAgg{
							ChartType: "line",
							Name:      name,
						},
					},
				},
			},
		},
	}
}

// LogStatisticResult .
type LogStatisticResult struct {
	Name string            `json:"name"`
	Data []*CountHistogram `json:"data"`
}

// CountHistogram .
type CountHistogram struct {
	Count ArrayAgg `json:"count"`
}

// ArrayAgg .
type ArrayAgg struct {
	UnitType  string    `json:"unitType"`
	Unit      string    `json:"unit"`
	ChartType string    `json:"chartType"`
	AxisIndex int64     `json:"axisIndex"`
	Name      string    `json:"name"`
	Tag       string    `json:"tag"`
	Data      []float64 `json:"data"`
}

func (c *ESClient) searchLogs(req *LogSearchRequest, timeout time.Duration) (*LogQueryResponse, error) {
	switch c.LogVersion {
	case LogVersion1:
		return c.searchLogsV1(req, timeout)
	}
	return c.searchLogsV2(req, timeout)
}

func (c *ESClient) statisticLogs(req *LogStatisticRequest, timeout time.Duration, name string) (*LogStatisticResponse, error) {
	switch c.LogVersion {
	case LogVersion1:
		return c.statisticLogsV1(req, timeout, name)
	}
	return c.statisticLogsV2(req, timeout, name)
}

func (c *ESClient) getTagsBoolQuery(req *LogRequest) *elastic.BoolQuery {
	boolQuery := elastic.NewBoolQuery()
	for _, item := range req.Filters {
		if item.Key != "origin" {
			boolQuery = boolQuery.Filter(elastic.NewTermQuery("tags."+item.Key, item.Value))
		}
	}
	if c.LogVersion != LogVersion1 {
		boolQuery.Filter(elastic.NewTermQuery("tags.dice_org_id", strconv.FormatInt(req.OrgID, 10)))
	}
	return boolQuery
}

func (c *ESClient) getSearchSource(req *LogSearchRequest, boolQuery *elastic.BoolQuery) *elastic.SearchSource {
	searchSource := elastic.NewSearchSource().Query(boolQuery)
	if len(req.Sort) > 0 {
		sorts := strings.Split(req.Sort, ",")
		for _, sort := range sorts {
			ascending := true
			parts := strings.SplitN(sort, " ", 2)
			if len(parts) > 1 && "desc" == strings.ToLower(strings.TrimSpace(parts[1])) {
				ascending = false
			}
			key := strings.TrimSpace(parts[0])
			if c.LogVersion == LogVersion1 && key == "timestamp" {
				key = "@timestamp"
			}
			searchSource.Sort(key, ascending)
		}
	}
	if req.Highlight {
		searchSource.Highlight(elastic.NewHighlight().
			PreTags("").
			PostTags("").
			FragmentSize(1).
			RequireFieldMatch(true).
			Field("*"))
	}
	searchSource.From(int((req.Page - 1) * req.Size)).Size(int(req.Size))
	return searchSource
}

func (c *ESClient) doRequest(req *LogRequest, searchSource *elastic.SearchSource, timeout time.Duration) (*elastic.SearchResult, error) {
	var indices []string
	if len(req.Addon) > 0 {
		start := req.Start * int64(time.Millisecond)
		end := req.End * int64(time.Millisecond)
		for _, entry := range c.Entrys {
			if (entry.MinTS == 0 || entry.MinTS <= end) &&
				(entry.MaxTS == 0 || entry.MaxTS >= start) {
				indices = append(indices, entry.Index)
			}
		}
		if req.Debug {
			fmt.Println(start, end, indices)
		}
	}
	if len(indices) <= 0 {
		indices = c.Indices
	}
	context, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resp, err := c.Client.Search(indices...).
		IgnoreUnavailable(true).
		AllowNoIndices(true).
		SearchSource(searchSource).Do(context)
	if err != nil || (resp != nil && resp.Error != nil) {
		if resp != nil && resp.Error != nil {
			return nil, fmt.Errorf("fail to request es: %s", jsonx.MarshalAndIndent(resp.Error))
		}
		return nil, fmt.Errorf("fail to request es: %s", err)
	}
	return resp, nil
}

func (c *ESClient) doSearchLogs(req *LogSearchRequest, searchSource *elastic.SearchSource, timeout time.Duration) (int64, []*elastic.SearchHit, error) {
	resp, err := c.doRequest(&req.LogRequest, searchSource, timeout)
	if err != nil {
		return 0, nil, err
	}
	if resp == nil || resp.Hits == nil || len(resp.Hits.Hits) <= 0 {
		return 0, nil, nil
	}
	return resp.TotalHits(), resp.Hits.Hits, nil
}

func (c *ESClient) setModule(log *logs.Log) {
	if log.Tags != nil {
		if log.Tags["origin"] == "sls" {
			project := log.Tags["sls_project"]
			logStore := log.Tags["sls_log_store"]
			if len(project) > 0 && len(logStore) > 0 {
				log.Tags["module"] = project + "/" + logStore
			}
		} else {
			project := log.Tags["dice_project_name"]
			app := log.Tags["dice_application_name"]
			service := log.Tags["dice_service_name"]
			if len(project) > 0 && len(app) > 0 && len(service) > 0 {
				log.Tags["origin"] = "dice"
				log.Tags["module"] = project + "/" + app + "/" + service
			}
		}
	}
}

// SearchLogs .
func (p *provider) SearchLogs(req *LogSearchRequest) (interface{}, error) {
	clients := p.getESClients(req.OrgID, &req.LogRequest)
	var results []*LogQueryResponse
	for _, client := range clients {
		result, err := client.searchLogs(req, p.C.Timeout)
		if err != nil {
			continue
		}
		results = append(results, result)
	}
	// multiple result set appear only at org-level search scenario
	// as we are going to remove the entry from cloud management page
	// it's okay to ignore the page size limitation
	return mergeLogSearch(0, results), nil
}

func mergeLogSearch(limit int, results []*LogQueryResponse) *LogQueryResponse {
	if len(results) <= 0 {
		return &LogQueryResponse{}
	} else if len(results) == 1 {
		return results[0]
	}
	resp := &LogQueryResponse{}
	for _, result := range results {
		resp.Total += result.Total
	}
	var count int
	for limit == 0 || count < limit {
		var min *LogItem
		var idx int
		for i, result := range results {
			if len(result.Data) <= 0 {
				continue
			}
			first := result.Data[0]
			if min == nil {
				min = first
				idx = i
				continue
			}
			if first.Source.Timestamp < min.Source.Timestamp || (first.Source.Timestamp == min.Source.Timestamp && first.Source.Offset < min.Source.Offset) {
				min = first
				idx = i
				continue
			}
		}
		if min == nil {
			break
		}
		results[idx].Data = results[idx].Data[1:]
		resp.Data = append(resp.Data, min)
		count++
	}
	return resp
}

// StatisticLogs .
func (p *provider) StatisticLogs(req *LogStatisticRequest) (interface{}, error) {
	clients := p.getESClients(req.OrgID, &req.LogRequest)
	var results []*LogStatisticResponse
	name := p.t.Text(req.Lang, "Count")
	for _, client := range clients {
		result, err := client.statisticLogs(req, p.C.Timeout, name)
		if err != nil {
			continue
		}
		results = append(results, result)
	}
	return mergeStatisticResponse(results), nil
}

func mergeStatisticResponse(results []*LogStatisticResponse) *LogStatisticResponse {
	if len(results) <= 0 {
		return nil
	} else if len(results) == 1 {
		return results[0]
	}
	first := results[0]
	list := first.Results[0].Data[0].Count.Data
	for _, result := range results[1:] {
		first.Total += result.Total
		for i, item := range result.Results[0].Data[0].Count.Data {
			if i < len(list) {
				list[i] += item
			} else {
				list = append(list, item)
			}
		}
	}
	first.Results[0].Data[0].Count.Data = list
	return first
}

func (p *provider) ListDefaultFields() []*LogField {
	return []*LogField{
		&LogField{FieldName: "source", SupportAggregation: false, Display: false},
		&LogField{FieldName: "id", SupportAggregation: false, Display: true},
		&LogField{FieldName: "stream", SupportAggregation: false, Display: false},
		&LogField{FieldName: "content", SupportAggregation: false, Display: true},
		&LogField{FieldName: "offset", SupportAggregation: false, Display: false},
		&LogField{FieldName: "timestamp", SupportAggregation: false, Display: false},
		&LogField{FieldName: "uniId", SupportAggregation: false, Display: false},
		&LogField{FieldName: "tags.origin", SupportAggregation: false, Display: false},
		&LogField{FieldName: "tags.dice_org_id", SupportAggregation: false, Display: false},
		&LogField{FieldName: "tags.dice_org_name", SupportAggregation: false, Display: false},
		&LogField{FieldName: "tags.dice_cluster_name", SupportAggregation: true, Display: false},
		&LogField{FieldName: "tags.dice_project_id", SupportAggregation: false, Display: false},
		&LogField{FieldName: "tags.dice_project_name", SupportAggregation: true, Display: false},
		&LogField{FieldName: "tags.dice_application_id", SupportAggregation: false, Display: false},
		&LogField{FieldName: "tags.dice_application_name", SupportAggregation: true, Display: true},
		&LogField{FieldName: "tags.dice_runtime_id", SupportAggregation: false, Display: false},
		&LogField{FieldName: "tags.dice_runtime_name", SupportAggregation: false, Display: false},
		&LogField{FieldName: "tags.dice_workspace", SupportAggregation: true, Display: true},
		&LogField{FieldName: "tags.dice_service_name", SupportAggregation: true, Display: true},
		&LogField{FieldName: "tags.pod_namespace", SupportAggregation: true, Display: true},
		&LogField{FieldName: "tags.pod_name", SupportAggregation: true, Display: true},
		&LogField{FieldName: "tags.container_name", SupportAggregation: true, Display: true},
		&LogField{FieldName: "tags.request-id", SupportAggregation: false, Display: true},
	}
}
