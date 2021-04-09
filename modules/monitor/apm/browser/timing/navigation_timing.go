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

	utils "github.com/erda-project/erda/modules/monitor/utils"
)

// NavigationTiming .
type NavigationTiming struct {
	LoadTime          int64
	ReadyStart        int64
	DomReadyTime      int64
	ScriptExecuteTime int64
	RequestTime       int64
	ResponseTime      int64
	InitDomTreeTime   int64
	LoadEventTime     int64
	UnloadEventTime   int64
	AppCacheTime      int64
	ConnectTime       int64
	LookupDomainTime  int64
	RedirectTime      int64
}

// ParseNavigationTiming .
func ParseNavigationTiming(nt string) *NavigationTiming {
	navigationTiming := NavigationTiming{}
	if nt == "" {
		return &navigationTiming
	}
	times := strings.Split(nt, ",")
	ts := make([]int64, 13)
	t := utils.ParseInt64WithRadix(times[0], 0, 36)
	ts[0] = t
	for i := 1; i < 13 && i < len(times); i++ {
		ts[i] = utils.ParseInt64WithRadix(times[i], 0, 36)
		if ts[i] > t {
			ts[i] = 0
		}
	}
	navigationTiming.LoadTime = t
	navigationTiming.ReadyStart = ts[1]
	navigationTiming.DomReadyTime = ts[2]
	navigationTiming.ScriptExecuteTime = ts[3]
	navigationTiming.RequestTime = ts[4]
	navigationTiming.ResponseTime = ts[5]
	navigationTiming.InitDomTreeTime = ts[6]
	navigationTiming.LoadEventTime = ts[7]
	navigationTiming.UnloadEventTime = ts[8]
	navigationTiming.AppCacheTime = ts[9]
	navigationTiming.ConnectTime = ts[10]
	navigationTiming.LookupDomainTime = ts[11]
	navigationTiming.RedirectTime = ts[12]
	return &navigationTiming
}
