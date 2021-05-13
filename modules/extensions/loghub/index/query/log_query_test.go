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
	"fmt"

	"github.com/recallsong/go-utils/encoding/jsonx"

	"github.com/erda-project/erda/modules/monitor/core/logs"
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
	fmt.Println(jsonx.MarshalAndIntend(result), len(result.Data))

}
