package query

import (
	"fmt"

	"github.com/erda-project/erda/modules/monitor/core/logs"
	"github.com/recallsong/go-utils/encoding/jsonx"
)

func Example_mergeLogSearch() {
	limit := 10
	results := []*LogQueryResponse{
		&LogQueryResponse{
			Total: 11,
			Data: []*logs.Log{
				&logs.Log{
					Content:   "1",
					Timestamp: 1,
				},
				&logs.Log{
					Content:   "3",
					Timestamp: 3,
					Offset:    1,
				},
				&logs.Log{
					Content:   "3",
					Timestamp: 3,
					Offset:    2,
				},
				&logs.Log{
					Content:   "5",
					Timestamp: 5,
					Offset:    1,
				},
				&logs.Log{
					Content:   "5",
					Timestamp: 5,
					Offset:    2,
				},
				&logs.Log{
					Content:   "6",
					Timestamp: 6,
				},
				&logs.Log{
					Content:   "7",
					Timestamp: 7,
				},
				&logs.Log{
					Content:   "8",
					Timestamp: 8,
				},
				&logs.Log{
					Content:   "9",
					Timestamp: 9,
				},
				&logs.Log{
					Content:   "10",
					Timestamp: 10,
				},
				&logs.Log{
					Content:   "11",
					Timestamp: 11,
				},
			},
		},
		&LogQueryResponse{
			Total: 10,
			Data: []*logs.Log{
				&logs.Log{
					Content:   "2",
					Timestamp: 2,
				},
				&logs.Log{
					Content:   "3",
					Timestamp: 3,
					Offset:    3,
				},
				&logs.Log{
					Content:   "4",
					Timestamp: 4,
					Offset:    1,
				},
				&logs.Log{
					Content:   "4",
					Timestamp: 4,
					Offset:    2,
				},
				&logs.Log{
					Content:   "4",
					Timestamp: 4,
					Offset:    3,
				},
				&logs.Log{
					Content:   "4",
					Timestamp: 4,
					Offset:    4,
				},
				&logs.Log{
					Content:   "4",
					Timestamp: 4,
					Offset:    5,
				},
				&logs.Log{
					Content:   "4",
					Timestamp: 4,
					Offset:    6,
				},
				&logs.Log{
					Content:   "4",
					Timestamp: 4,
					Offset:    7,
				},
				&logs.Log{
					Content:   "4",
					Timestamp: 4,
					Offset:    7,
				},
			},
		},
	}
	result := mergeLogSearch(limit, results)
	fmt.Println(jsonx.MarshalAndIntend(result), len(result.Data))

	// Output:
	// TODO .
}
