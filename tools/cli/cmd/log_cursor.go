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

package cmd

import (
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
)

type logCursor struct {
	Start int64
}

type logInitialFetcher func() ([]apistructs.DashboardSpotLogLine, error)
type logPageFetcher func(start int64, count int64) ([]apistructs.DashboardSpotLogLine, error)

func fetchWatchLogLines(cursor logCursor, tail int, fetchInitial logInitialFetcher, fetchPage logPageFetcher) ([]apistructs.DashboardSpotLogLine, logCursor, error) {
	if cursor.Start <= 0 {
		lines, err := fetchInitial()
		if err != nil {
			return nil, cursor, err
		}
		return lines, advanceLogCursor(cursor, lines), nil
	}

	lines, err := fetchIncrementalLogLines(cursor.Start, defaultLogPageSize(tail, 0), fetchPage)
	if err != nil {
		return nil, cursor, err
	}
	return lines, advanceLogCursor(cursor, lines), nil
}

func fetchIncrementalLogLines(start int64, pageSize int64, fetchPage logPageFetcher) ([]apistructs.DashboardSpotLogLine, error) {
	pageSize = defaultLogPageSize(0, pageSize)

	var lines []apistructs.DashboardSpotLogLine
	for {
		page, err := fetchPage(start, pageSize)
		if err != nil {
			return nil, err
		}
		if len(page) == 0 {
			return lines, nil
		}

		lines = append(lines, page...)
		if int64(len(page)) < pageSize {
			return lines, nil
		}

		nextStart := nextLogStart(start, page)
		if nextStart <= start {
			return lines, nil
		}
		start = nextStart
	}
}

func defaultLogPageSize(tail int, count int64) int64 {
	if count > 0 {
		return count
	}
	return int64(normalizedLogTail(tail))
}

func advanceLogCursor(cursor logCursor, lines []apistructs.DashboardSpotLogLine) logCursor {
	nextStart := nextLogStart(cursor.Start, lines)
	if nextStart > cursor.Start {
		cursor.Start = nextStart
	}
	return cursor
}

func nextLogStart(currentStart int64, lines []apistructs.DashboardSpotLogLine) int64 {
	maxTimestamp := currentStart
	for _, line := range lines {
		ts, ok := parseLogTimestamp(line.TimeStamp)
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
	return maxTimestamp + 1
}

func parseLogTimestamp(value string) (int64, bool) {
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
