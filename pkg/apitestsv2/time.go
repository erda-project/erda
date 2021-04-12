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

package apitestsv2

import (
	"strconv"
	"time"
)

const (
	HourStampS  = 3600
	HourStampMs = 3600000
	HourStampNs = 3600000000000
)

const (
	TimeStamp            = "timestamp"               // 当前时间戳s格式：1587631494(s)
	TimeStampHour        = "timestamp_hour"          // 1小时前时间戳s格式：1587627894(s)
	TimeStampAfterHour   = "timestamp_after_hour"    // 1小时后时间戳s格式：1587636854(s)
	TimeStampDay         = "timestamp_day"           // 1天前时间戳s格式：1587545094(s)
	TimeStampAfterDay    = "timestamp_after_day"     // 1天后时间戳s格式：1587719654(s)
	TimeStampMs          = "timestamp_ms"            // 当前时间戳ms格式：1587631494093(ms)
	TimeStampMsHour      = "timestamp_ms_hour"       // 1小时前时间戳ms格式：1587627894093(ms)
	TimeStampMsAfterHour = "timestamp_ms_after_hour" // 1小时后时间戳ms格式：1587636854150(ms)
	TimeStampMsDay       = "timestamp_ms_day"        // 1天前时间戳ms格式：1587545094093(ms)
	TimeStampMsAfterDay  = "timestamp_ms_after_day"  // 1天后时间戳ms格式：1587719654150(ms)
	TimeStampNs          = "timestamp_ns"            // 当前时间戳ns格式：1587631494093562000(ns)
	TimeStampNsHour      = "timestamp_ns_hour"       // 1小时前时间戳ns格式：1587627894093566000(ns)
	TimeStampNsAfterHour = "timestamp_ns_after_hour" // 1小时后时间戳ns格式：1587636854150072000(ns)
	TimeStampNsDay       = "timestamp_ns_day"        // 1天前前时间戳ns格式：1587545094093570000(ns)
	TimeStampNsAfterDay  = "timestamp_ns_after_day"  // 1天前后时间戳ns格式：1587719654150080000(ns)
	Date                 = "date"                    // 当前日期格式：2006-01-02
	DateDay              = "date_day"                // 1天前日期格式：2006-01-01
	DateTime             = "datetime"                // 当前带时间的格式：2006-01-02 15:04:05
	DateTimeHour         = "datetime_hour"           // 1小时前带时间的格式：2006-01-02 14:04:05
)

func getTime(timeType string) string {
	hour, _ := time.ParseDuration("-1h")
	day, _ := time.ParseDuration("-24h")
	currentTime := time.Now()
	switch timeType {
	case TimeStamp:
		return strconv.FormatInt(currentTime.Unix(), 10)
	case TimeStampHour:
		return strconv.FormatInt(currentTime.Unix()-HourStampS, 10)
	case TimeStampAfterHour:
		return strconv.FormatInt(currentTime.Unix()+HourStampS, 10)
	case TimeStampDay:
		return strconv.FormatInt(currentTime.Unix()-HourStampS*24, 10)
	case TimeStampAfterDay:
		return strconv.FormatInt(currentTime.Unix()+HourStampS*24, 10)
	case TimeStampMs:
		return strconv.FormatInt(currentTime.UnixNano()/1e6, 10)
	case TimeStampMsHour:
		return strconv.FormatInt(currentTime.UnixNano()/1e6-HourStampMs, 10)
	case TimeStampMsAfterHour:
		return strconv.FormatInt(currentTime.UnixNano()/1e6+HourStampMs, 10)
	case TimeStampMsDay:
		return strconv.FormatInt(currentTime.UnixNano()/1e6-HourStampMs*24, 10)
	case TimeStampMsAfterDay:
		return strconv.FormatInt(currentTime.UnixNano()/1e6+HourStampMs*24, 10)
	case TimeStampNs:
		return strconv.FormatInt(currentTime.UnixNano(), 10)
	case TimeStampNsHour:
		return strconv.FormatInt(currentTime.UnixNano()-HourStampNs, 10)
	case TimeStampNsAfterHour:
		return strconv.FormatInt(currentTime.UnixNano()+HourStampNs, 10)
	case TimeStampNsDay:
		return strconv.FormatInt(currentTime.UnixNano()-HourStampNs*24, 10)
	case TimeStampNsAfterDay:
		return strconv.FormatInt(currentTime.UnixNano()+HourStampNs*24, 10)
	case Date:
		return currentTime.Format("2006-01-02")
	case DateDay:
		return currentTime.Add(day).Format("2006-01-02")
	case DateTime:
		return currentTime.Format("2006-01-02 15:04:05")
	case DateTimeHour:
		return currentTime.Add(hour).Format("2006-01-02 15:04:05")
	}

	return ""
}
