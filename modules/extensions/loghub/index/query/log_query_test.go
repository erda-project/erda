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
	"fmt"
	"testing"

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
	p := &provider{}

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
