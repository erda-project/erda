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

package common

import (
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
)

type LogCursor struct {
	Start int64
}

type LogInitialFetcher func() ([]apistructs.DashboardSpotLogLine, error)
type LogPageFetcher func(start int64, count int64) ([]apistructs.DashboardSpotLogLine, error)

const maxLogPageSize = int64(5000)

func FetchWatchLogLines(cursor LogCursor, tail int, fetchInitial LogInitialFetcher, fetchPage LogPageFetcher) ([]apistructs.DashboardSpotLogLine, LogCursor, error) {
	if cursor.Start <= 0 {
		lines, err := fetchInitial()
		if err != nil {
			return nil, cursor, err
		}
		return lines, AdvanceLogCursor(cursor, lines), nil
	}

	lines, err := FetchIncrementalLogLines(cursor.Start, DefaultLogPageSize(tail, 0), fetchPage)
	if err != nil {
		return nil, cursor, err
	}
	return lines, AdvanceLogCursor(cursor, lines), nil
}

func FetchIncrementalLogLines(start int64, pageSize int64, fetchPage LogPageFetcher) ([]apistructs.DashboardSpotLogLine, error) {
	basePageSize := DefaultLogPageSize(0, pageSize)
	pageSize = basePageSize

	var lines []apistructs.DashboardSpotLogLine
	for {
		requestStart := queryLogStart(start)
		var page []apistructs.DashboardSpotLogLine
		for {
			var err error
			page, err = fetchPage(requestStart, pageSize)
			if err != nil {
				return nil, err
			}
			if len(page) == 0 {
				return lines, nil
			}

			if shouldExpandLogPage(start, page, pageSize) {
				nextPageSize := expandLogPageSize(pageSize)
				if nextPageSize > pageSize {
					pageSize = nextPageSize
					continue
				}
			}
			break
		}

		lines = append(lines, page...)
		if int64(len(page)) < pageSize {
			return lines, nil
		}

		nextStart := NextLogStart(start, page)
		if nextStart <= start {
			return lines, nil
		}
		start = nextStart
		pageSize = basePageSize
	}
}

func DefaultLogPageSize(tail int, count int64) int64 {
	if count > 0 {
		return count
	}
	if tail <= 0 {
		return 200
	}
	return int64(tail)
}

func AdvanceLogCursor(cursor LogCursor, lines []apistructs.DashboardSpotLogLine) LogCursor {
	nextStart := NextLogStart(cursor.Start, lines)
	if nextStart > cursor.Start {
		cursor.Start = nextStart
	}
	return cursor
}

func NextLogStart(currentStart int64, lines []apistructs.DashboardSpotLogLine) int64 {
	maxTimestamp := currentStart
	for _, line := range lines {
		ts, ok := ParseLogTimestamp(line.TimeStamp)
		if !ok {
			continue
		}
		if ts > maxTimestamp {
			maxTimestamp = ts
		}
	}
	if maxTimestamp == currentStart {
		return currentStart
	}
	return maxTimestamp
}

func ParseLogTimestamp(value string) (int64, bool) {
	if value == "" {
		return 0, false
	}
	if ts, err := strconv.ParseInt(value, 10, 64); err == nil {
		return ts, true
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if ts, err := time.Parse(layout, value); err == nil {
			return ts.UnixNano(), true
		}
	}
	return 0, false
}

func queryLogStart(cursorStart int64) int64 {
	if cursorStart <= 0 {
		return 0
	}
	return cursorStart - 1
}

func shouldExpandLogPage(cursorStart int64, lines []apistructs.DashboardSpotLogLine, pageSize int64) bool {
	return cursorStart > 0 && int64(len(lines)) >= pageSize && NextLogStart(cursorStart, lines) <= cursorStart
}

func expandLogPageSize(pageSize int64) int64 {
	if pageSize >= maxLogPageSize {
		return pageSize
	}
	next := pageSize * 2
	if next > maxLogPageSize {
		return maxLogPageSize
	}
	return next
}
