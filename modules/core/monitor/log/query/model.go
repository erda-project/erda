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
