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

	"github.com/recallsong/go-utils/encoding/jsonx"

	logs "github.com/erda-project/erda/modules/core/monitor/log"
)

func Example_mergeLogSearch() {
	limit := 10
	results := []*LogQueryResponse{
		{
			Total: 11,
			Data: []*logs.Log{
				{
					Content:   "1",
					Timestamp: 1,
				},
				{
					Content:   "3",
					Timestamp: 3,
					Offset:    1,
				},
				{
					Content:   "3",
					Timestamp: 3,
					Offset:    2,
				},
				{
					Content:   "5",
					Timestamp: 5,
					Offset:    1,
				},
				{
					Content:   "5",
					Timestamp: 5,
					Offset:    2,
				},
				{
					Content:   "6",
					Timestamp: 6,
				},
				{
					Content:   "7",
					Timestamp: 7,
				},
				{
					Content:   "8",
					Timestamp: 8,
				},
				{
					Content:   "9",
					Timestamp: 9,
				},
				{
					Content:   "10",
					Timestamp: 10,
				},
				{
					Content:   "11",
					Timestamp: 11,
				},
			},
		},
		{
			Total: 10,
			Data: []*logs.Log{
				{
					Content:   "2",
					Timestamp: 2,
				},
				{
					Content:   "3",
					Timestamp: 3,
					Offset:    3,
				},
				{
					Content:   "4",
					Timestamp: 4,
					Offset:    1,
				},
				{
					Content:   "4",
					Timestamp: 4,
					Offset:    2,
				},
				{
					Content:   "4",
					Timestamp: 4,
					Offset:    3,
				},
				{
					Content:   "4",
					Timestamp: 4,
					Offset:    4,
				},
				{
					Content:   "4",
					Timestamp: 4,
					Offset:    5,
				},
				{
					Content:   "4",
					Timestamp: 4,
					Offset:    6,
				},
				{
					Content:   "4",
					Timestamp: 4,
					Offset:    7,
				},
				{
					Content:   "4",
					Timestamp: 4,
					Offset:    7,
				},
			},
		},
	}
	result := mergeLogSearch(limit, results)
	fmt.Println(jsonx.MarshalAndIndent(result), len(result.Data))

}
