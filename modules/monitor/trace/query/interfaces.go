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
	"github.com/gocql/gocql"
)

type (
	SpanQueryAPI interface {
		SelectSpans(traceId string, limit int64) []map[string]interface{}
	}
)

func (p *provider) SelectSpans(traceId string, limit int64) []map[string]interface{} {
	return p.spansResult(traceId, limit)
}

func (p *provider) selectSpans(traceId string, limit int64) *gocql.Iter {
	return p.cassandraSession.Query("SELECT * FROM spans WHERE trace_id = ? limit ?", traceId, limit).
		Consistency(gocql.All).RetryPolicy(nil).Iter()
}

func (p *provider) spansResult(traceId string, limit int64) []map[string]interface{} {
	iter := p.selectSpans(traceId, limit)
	list := make([]map[string]interface{}, 0, 10)
	for row := make(map[string]interface{}, 0); iter.MapScan(row); {
		list = append(list, row)
	}
	return list
}
