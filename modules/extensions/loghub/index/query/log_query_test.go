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
	"fmt"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/encoding/jsonx"

	logs "github.com/erda-project/erda/modules/core/monitor/log"
)

func Example_mergeLogSearch() {
	limit := 10
	results := []*LogQueryResponse{
		{
			Total: 11,
			Data: []*LogItem{
				{
					Source: &logs.Log{
						Content:   "1",
						Timestamp: 1,
					},
				},
				{
					Source: &logs.Log{
						Content:   "3",
						Timestamp: 3,
						Offset:    1,
					},
				},
				{
					Source: &logs.Log{
						Content:   "3",
						Timestamp: 3,
						Offset:    2,
					},
				},
				{
					Source: &logs.Log{
						Content:   "5",
						Timestamp: 5,
						Offset:    1,
					},
				},
				{
					Source: &logs.Log{
						Content:   "5",
						Timestamp: 5,
						Offset:    2,
					},
				},
				{
					Source: &logs.Log{
						Content:   "6",
						Timestamp: 6,
					},
				},
				{
					Source: &logs.Log{
						Content:   "7",
						Timestamp: 7,
					},
				},
				{
					Source: &logs.Log{
						Content:   "8",
						Timestamp: 8,
					},
				},
				{},
				{
					Source: &logs.Log{
						Content:   "10",
						Timestamp: 10,
					},
				},
				{
					Source: &logs.Log{
						Content:   "11",
						Timestamp: 11,
					},
				},
			},
		},
		{
			Total: 10,
			Data: []*LogItem{
				{
					Source: &logs.Log{
						Content:   "2",
						Timestamp: 2,
					},
				},
				{
					Source: &logs.Log{
						Content:   "3",
						Timestamp: 3,
						Offset:    3,
					},
				},
				{
					Source: &logs.Log{
						Content:   "4",
						Timestamp: 4,
						Offset:    1,
					},
				},
				{
					Source: &logs.Log{
						Content:   "4",
						Timestamp: 4,
						Offset:    2,
					},
				},
				{
					Source: &logs.Log{
						Content:   "4",
						Timestamp: 4,
						Offset:    3,
					},
				},
				{
					Source: &logs.Log{
						Content:   "4",
						Timestamp: 4,
						Offset:    4,
					},
				},
				{
					Source: &logs.Log{
						Content:   "4",
						Timestamp: 4,
						Offset:    5,
					},
				},
				{
					Source: &logs.Log{
						Content:   "4",
						Timestamp: 4,
						Offset:    6,
					},
				},
				{
					Source: &logs.Log{
						Content:   "4",
						Timestamp: 4,
						Offset:    7,
					},
				},
				{
					Source: &logs.Log{
						Content:   "4",
						Timestamp: 4,
						Offset:    7,
					},
				},
			},
		},
	}
	result := mergeLogSearch(limit, results)
	fmt.Println(jsonx.MarshalAndIndent(result), len(result.Data))

}

func TestListDefaultFields_Should_Success(t *testing.T) {
	p := &provider{
		C: &config{
			IndexFieldSettings: struct {
				File            string               `file:"file"`
				DefaultSettings defaultFieldSettings `file:"default_settings"`
			}{
				File: "",
				DefaultSettings: defaultFieldSettings{
					Fields: []logField{
						{
							AllowEdit:          true,
							FieldName:          "field-1",
							Display:            true,
							SupportAggregation: true,
						},
					},
				},
			},
		},
	}

	result := p.ListDefaultFields()
	if len(result) == 0 {
		t.Errorf("should not return empty slice")
	}
}

func Test_concatBucketSlices(t *testing.T) {
	limit := 10
	slices := [][]*BucketAgg{
		{
			{Key: "4", Count: 5},
			{Key: "1", Count: 2},
			{Key: "3", Count: 2},
		},
		{
			{Key: "2", Count: 11},
			{Key: "3", Count: 4},
		},
	}

	want := []*BucketAgg{
		{Key: "2", Count: 11},
		{Key: "3", Count: 6},
		{Key: "4", Count: 5},
		{Key: "1", Count: 2},
	}

	result := concatBucketSlices(limit, slices...)

	if len(result) != len(want) {
		t.Errorf("same key should merged")
	}

	for i, agg := range result {
		if agg.Key != want[i].Key || agg.Count != want[i].Count {
			t.Errorf("expect key: %s count: %d, but got key: %s count: %d", want[i].Key, want[i].Count, agg.Key, agg.Count)
		}
	}
}

func Test_getSearchSource_Should_Sort_As_Expect(t *testing.T) {
	c := &ESClient{}
	req := &LogSearchRequest{
		Sort: []string{"timestamp desc", "offset desc"},
	}
	result, err := c.getSearchSource(req, elastic.NewBoolQuery()).Source()
	if err != nil {
		t.Errorf("should not error getting serialized search source")
	}
	data := fmt.Sprintf("%+v", result.(map[string]interface{})["sort"])
	expect := "[map[timestamp:map[order:desc]] map[offset:map[order:desc]]]"
	if data != expect {
		t.Errorf("sort assert failed, expect: %s, but got: %s", expect, data)
	}
}

func Test_getSearchSource_Should_Include_SearchAfter(t *testing.T) {
	c := &ESClient{}
	req := &LogSearchRequest{
		SearchAfter: []interface{}{"12343434", 123, 123},
	}
	result, err := c.getSearchSource(req, elastic.NewBoolQuery()).Source()
	if err != nil {
		t.Errorf("should not error getting serialized search source")
	}
	data := result.(map[string]interface{})["search_after"].([]interface{})
	if len(data) != len(req.SearchAfter) {
		t.Errorf("search_after generated not as expect, expect len: %d, but got len: %d", len(req.SearchAfter), len(data))
	}
	for i, item := range data {
		if item != req.SearchAfter[i] {
			t.Errorf("search_after generated not as expect")
		}
	}
}

func Test_aggregateFields_With_ValidParams_Should_Success(t *testing.T) {
	c := &ESClient{}
	want := struct {
		Total     int64
		AggFields map[string]*LogFieldBucket
	}{
		Total: 10,
		AggFields: map[string]*LogFieldBucket{
			"tags.dice_application_name": {
				Buckets: []*BucketAgg{
					{
						Key:   "app-1",
						Count: 10,
					},
				},
			},
		},
	}

	defer monkey.Unpatch((*ESClient).doRequest)
	monkey.Patch((*ESClient).doRequest, func(client *ESClient, req *LogRequest, searchSource *elastic.SearchSource, timeout time.Duration) (*elastic.SearchResult, error) {

		type AggregationBucketKeyItem struct {
			Key         interface{} `json:"key"`
			KeyAsString *string     `json:"key_as_string"`
			KeyNumber   json.Number
			DocCount    int64 `json:"doc_count"`
		}

		type AggregationBucketKeyItems struct {
			DocCountErrorUpperBound int64                       `json:"doc_count_error_upper_bound"`
			SumOfOtherDocCount      int64                       `json:"sum_other_doc_count"`
			Buckets                 []*AggregationBucketKeyItem `json:"buckets"`
			Meta                    map[string]interface{}      `json:"meta,omitempty"`
		}

		aggs := elastic.Aggregations{}
		for key, bucket := range want.AggFields {

			var buckets []*AggregationBucketKeyItem
			for _, b := range bucket.Buckets {
				buckets = append(buckets, &AggregationBucketKeyItem{
					Key:      b.Key,
					DocCount: b.Count,
				})
			}

			bytes, _ := json.Marshal(AggregationBucketKeyItems{
				Buckets: buckets,
			})
			rawMessage := json.RawMessage(bytes)
			aggs[key] = &rawMessage
		}

		return &elastic.SearchResult{
			Hits: &elastic.SearchHits{
				TotalHits: want.Total,
			},
			Aggregations: aggs,
		}, nil
	})

	req := &LogFieldsAggregationRequest{
		LogRequest: LogRequest{
			OrgID:       1,
			ClusterName: "cluster-1",
			Addon:       "addon-1",
			Start:       time.Now().Unix(),
			End:         time.Now().Unix(),
		},
		TermsSize: 10,
		AggFields: []string{"tags.dice_application_name"},
	}

	result, err := c.aggregateFields(req, time.Minute)
	if err != nil {
		t.Errorf("aggregateFields with valid params should not error: %+v", err)
	}

	if result.Total != want.Total {
		t.Errorf("total count assert failed, expect :%d, but got: %d", want.Total, result.Total)
	}

	for key, expectBuckets := range want.AggFields {
		resultBuckets, ok := result.AggFields[key]
		if !ok {
			t.Errorf("assert aggField failed, key [%s] not exists", key)
		}
		if len(resultBuckets.Buckets) != len(expectBuckets.Buckets) {
			t.Errorf("assert aggField failed, bucket len not match")
		}
	}
}
