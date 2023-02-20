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

package loader

import (
	"fmt"
	"strconv"
	"strings"
)

type TableMeta struct {
	CreateTableSQL string
	Engine         string
	Columns        map[string]*TableColumn
	TTLDays        int64
	HotTTLDays     int64
	TTLBaseField   string
	TimeKey        string
}

func (meta *TableMeta) HasColdHotTTL() bool {
	if len(meta.TTLBaseField) < 0 {
		return false
	}
	return strings.Index(meta.TTLBaseField, "slow") != -1
}

func (meta *TableMeta) extractTTLDays() {
	// 1. extract ttl ~ settings
	ttlText := GetStringInBetween(meta.CreateTableSQL, "TTL", "SETTINGS")
	if len(ttlText) <= 0 {
		return
	}
	//toDateTime(timestamp) + toIntervalHour(1) TO VOLUME 'slow', toDateTime(timestamp) + toIntervalDay(7)
	// 2. split ttl
	ttls := strings.Split(ttlText, ",")
	if len(ttls) <= 0 {
		return
	}
	// 3. extract last ttl day
	last := ttls[len(ttls)-1]
	ttlDays := GetStringInBetween(last, "toIntervalDay(", ")")
	meta.TTLDays, _ = strconv.ParseInt(ttlDays, 10, 64)

	meta.TimeKey = last[:strings.Index(last, "+")]
	meta.TimeKey = strings.TrimSpace(meta.TimeKey)

	// 4. extract hot ttl
	hot := ttls[0]
	if s := strings.Index(hot, "TO VOLUME 'slow'"); s != -1 {
		hotDays := GetStringInBetween(hot, "toIntervalDay(", ")")
		hotTTL, _ := strconv.ParseInt(hotDays, 10, 64)
		meta.HotTTLDays = hotTTL
		fmt.Println(hotTTL)
	}

}

func GetStringInBetween(str string, left, right string) string {
	s := strings.Index(str, left)
	if s == -1 {
		return ""
	}
	s += len(left)
	e := strings.Index(str[s:], right)
	if e == -1 {
		return ""
	}
	return strings.TrimSpace(str[s : s+e])
}
