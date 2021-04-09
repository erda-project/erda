// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package timing

import (
	"strings"

	"github.com/erda-project/erda/modules/monitor/utils"
)

// PerformanceTiming .
type PerformanceTiming struct {
	FirstPaintTime             int64
	ConnectEnd                 int64
	ConnectStart               int64
	DomComplete                int64
	DomContentLoadedEventEnd   int64
	DomContentLoadedEventStart int64
	DomInteractive             int64
	DomLoading                 int64
	DomainLookupEnd            int64
	DomainLookupStart          int64
	FetchStart                 int64
	LoadEventEnd               int64
	LoadEventStart             int64
	NavigationStart            int64
	RedirectEnd                int64
	RedirectStart              int64
	RequestStart               int64
	ResponseEnd                int64
	ResponseStart              int64
	SecureConnectionStart      int64
	UnloadEventEnd             int64
	UnloadEventStart           int64
}

// ParsePerformanceTiming .
func ParsePerformanceTiming(pt string) *PerformanceTiming {
	performanceTiming := PerformanceTiming{}
	if pt == "" {
		return &performanceTiming
	}

	times := strings.Split(pt, ",")
	ts := make([]int64, 22)
	for i := 0; i < 22 && i < len(times); i++ {
		ts[i] = utils.ParseInt64WithRadix(times[i], 0, 36)
	}
	performanceTiming.FirstPaintTime = ts[0]
	if performanceTiming.FirstPaintTime < 0 {
		performanceTiming.FirstPaintTime = 0
	}

	performanceTiming.ConnectEnd = ts[1]
	performanceTiming.ConnectStart = ts[2]
	performanceTiming.DomComplete = ts[3]
	performanceTiming.DomContentLoadedEventEnd = ts[4]
	performanceTiming.DomContentLoadedEventStart = ts[5]
	performanceTiming.DomInteractive = ts[6]
	performanceTiming.DomLoading = ts[7]
	performanceTiming.DomainLookupEnd = ts[8]
	performanceTiming.DomainLookupStart = ts[9]
	performanceTiming.FetchStart = ts[10]
	performanceTiming.LoadEventEnd = ts[11]
	performanceTiming.LoadEventStart = ts[12]
	performanceTiming.NavigationStart = ts[13]
	performanceTiming.RedirectEnd = ts[14]
	performanceTiming.RedirectStart = ts[15]
	performanceTiming.RequestStart = ts[16]
	performanceTiming.ResponseEnd = ts[17]
	performanceTiming.ResponseStart = ts[18]
	performanceTiming.SecureConnectionStart = ts[19]
	performanceTiming.UnloadEventEnd = ts[20]
	performanceTiming.UnloadEventStart = ts[21]

	return &performanceTiming
}
