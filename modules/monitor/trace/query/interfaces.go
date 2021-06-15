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
	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
	"github.com/gocql/gocql"
)

type (
	SpanQueryAPI interface {
		SelectSpans(traceId string, limit int64) []*pb.Span
	}
)

func (p *provider) SelectSpans(traceId string, limit int64) []*pb.Span {
	return p.spansResult(traceId, limit)
}

func (p *provider) selectSpans(traceId string, limit int64) *gocql.Iter {
	return p.cassandraSession.Query("SELECT * FROM spans WHERE trace_id = ? limit ?", traceId, limit).Consistency(gocql.All).Iter()
}

func (p *provider) spansResult(traceId string, limit int64) []*pb.Span {
	iter := p.selectSpans(traceId, limit)
	var spans []*pb.Span
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		var span pb.Span
		span.Id = row["span_id"].(string)
		span.TraceId = row["trace_id"].(string)
		span.OperationName = row["operation_name"].(string)
		span.ParentSpanId = row["parent_span_id"].(string)
		span.StartTime = row["start_time"].(int64)
		span.EndTime = row["end_time"].(int64)
		span.Tags = row["tags"].(map[string]string)
		spans = append(spans, &span)
	}
	return spans
}
