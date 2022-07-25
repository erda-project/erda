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

package model

import (
	"bytes"
	"strings"
	"time"
)

// Data .
type Data struct {
	Total    int64
	Interval int64
	Columns  []*Column
	Rows     [][]interface{}
}

// Column .
type Column struct {
	Type  string
	Name  string
	Key   string
	Flag  ColumnFlag
	Extra interface{}
}

// ColumnFlag .
type ColumnFlag int32

// ColumnFlag values
const (
	ColumnFlagNone = ColumnFlag(0)
	ColumnFlagHide = ColumnFlag(1 << (iota - 1))
	ColumnFlagName
	ColumnFlagTimestamp
	ColumnFlagTag
	ColumnFlagField
	ColumnFlagLiteral
	ColumnFlagFunc
	ColumnFlagAgg
	ColumnFlagGroupBy
	ColumnFlagOrderBy

	ColumnFlagGroupByInterval
	ColumnFlagGroupByRange
)

func (f ColumnFlag) String() string {
	if f == ColumnFlagNone {
		return "node"
	}
	buf := &bytes.Buffer{}
	if f&ColumnFlagHide != 0 {
		buf.WriteString("hide|")
	}
	if f&ColumnFlagName != 0 {
		buf.WriteString("name|")
	}
	if f&ColumnFlagTimestamp != 0 {
		buf.WriteString("timestamp|")
	}
	if f&ColumnFlagTag != 0 {
		buf.WriteString("tag|")
	}
	if f&ColumnFlagField != 0 {
		buf.WriteString("field|")
	}
	if f&ColumnFlagLiteral != 0 {
		buf.WriteString("literal|")
	}
	if f&ColumnFlagFunc != 0 {
		buf.WriteString("func|")
	}
	if f&ColumnFlagAgg != 0 {
		buf.WriteString("agg|")
	}
	if f&ColumnFlagGroupBy != 0 {
		buf.WriteString("groupby")
		if f&(ColumnFlagGroupByInterval|ColumnFlagGroupByRange) != 0 {
			buf.WriteString(":")
			var group []string
			if f&ColumnFlagGroupByInterval != 0 {
				group = append(group, "interval")
			}
			if f&ColumnFlagGroupByRange != 0 {
				group = append(group, "range")
			}
			buf.WriteString(strings.Join(group, ","))
		}
		buf.WriteString("|")
	}
	if f&ColumnFlagOrderBy != 0 {
		buf.WriteString("orderby|")
	}
	return string(buf.Bytes()[0 : buf.Len()-1])
}

// Source .
type Source struct {
	Database string
	Name     string
}

// ResultSet .
type ResultSet struct {
	*Data
	Details interface{}
	Elapsed struct {
		Search time.Duration `json:"search"`
	} `json:"elapsed"`
}
