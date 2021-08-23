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

package query

import (
	"github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
)

// LogTableName .
const (
	LogMetaTableName = "spot_prod.base_log_meta"
)

// Logs .
type Logs []*pb.LogItem

func (l Logs) Len() int           { return len(l) }
func (l Logs) Less(i, j int) bool { return l[i].Timestamp < l[j].Timestamp }
func (l Logs) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

type LogMeta struct {
	Source string            `json:"source"`
	ID     string            `json:"id"`
	Tags   map[string]string `json:"tags"`
}

// SavedLog Cassandra查询的结构
type SavedLog struct {
	Source     string
	ID         string
	Stream     string
	TimeBucket int64 `db:"time_bucket"`
	Timestamp  int64
	Offset     int64
	Content    []byte
	Level      string
	RequestID  string `db:"request_id"`
}

type SaveLogMeta struct {
	Source string
	ID     string
	Tags   map[string]string
}
