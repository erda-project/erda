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
//
// Data format in elasticsearch
// {
// 	"name":"table_name",
//     "tags":{
//         "tag1":"val1",
//         "tag2":"val2"
//     },
//     "fields":{
//         "field1":1,
//         "field2":2
//     },
//     "timestamp":1599551100000000000
// }

package tsql

import (
	"context"
	"fmt"
	"time"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/meta"
)

// TimeUnit .
type TimeUnit time.Duration

// TimeUnit values
const (
	UnsetTimeUnit = TimeUnit(0)
	Nanosecond    = TimeUnit(time.Nanosecond)
	Microsecond   = TimeUnit(time.Microsecond)
	Millisecond   = TimeUnit(time.Millisecond)
	Second        = TimeUnit(time.Second)
	Minute        = TimeUnit(time.Minute)
	Hour          = TimeUnit(time.Hour)
	Day           = TimeUnit(24 * time.Hour)
)

// ParseTimeUnit .
func ParseTimeUnit(u string) (TimeUnit, error) {
	switch u {
	case "ns":
		return Nanosecond, nil
	case "us":
		return Microsecond, nil
	case "µs":
		return Microsecond, nil
	case "μs":
		return Microsecond, nil
	case "ms":
		return Millisecond, nil
	case "s":
		return Second, nil
	case "m":
		return Minute, nil
	case "h":
		return Hour, nil
	case "day":
		return Day, nil
	}
	return UnsetTimeUnit, fmt.Errorf("invaild time unit '%s'", u)
}

// ConvertTimestamp .
func ConvertTimestamp(t int64, from, to TimeUnit) int64 {
	if to == UnsetTimeUnit {
		return t
	}
	if from == UnsetTimeUnit {
		from = Nanosecond
	}
	return t * int64(from) / int64(to)
}

// Query .
type Query interface {
	Sources() []*model.Source
	SearchSource() interface{}
	SubSearchSource() interface{}
	AppendBoolFilter(key string, value interface{})
	ParseResult(ctx context.Context, resp interface{}) (*model.Data, error)
	Context() Context
	Debug() bool
	Timestamp() (int64, int64)
	Kind() string
	OrgName() string
	TerminusKey() string
}

// ErrNotSupportNonQueryStatement .
var ErrNotSupportNonQueryStatement = fmt.Errorf("not support non query statement")

// Parser .
type Parser interface {
	SetParams(params map[string]interface{}) Parser
	SetFilter(filter []*model.Filter) (Parser, error)
	SetOriginalTimeUnit(unit TimeUnit) Parser
	SetTargetTimeUnit(unit TimeUnit) Parser
	SetTimeKey(key string) Parser
	SetMaxTimePoints(points int64) Parser
	ParseQuery(ctx context.Context, kind string) ([]Query, error)
	Build() error
	Metrics() ([]string, error)
	SetOrgName(org string) Parser
	GetOrgName() string
	GetTerminusKey() string
	SetTerminusKey(terminusKey string) Parser
	SetMeta(meta.Interface)
}

// Creator .
type Creator func(start, end int64, stmt string, debug bool) Parser

// Parsers .
var Parsers = map[string]Creator{}

// RegisterParser .
func RegisterParser(name string, c Creator) {
	Parsers[name] = c
}

// New .
func New(start, end int64, ql, stmt string, debug bool) Parser {
	creator := Parsers[ql]
	if creator != nil {
		return creator(start, end, stmt, debug)
	}
	return nil
}
