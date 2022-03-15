package source

import (
	"context"

	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
)

type CassandraSource struct {
	CassandraSession *cassandra.Session
}

func (cs *CassandraSource) GetSpans(ctx context.Context, req *pb.GetSpansRequest) []*pb.Span {
	iter := cs.CassandraSession.Session().Query("SELECT * FROM spans WHERE trace_id = ? limit ?", req.TraceID, req.Limit).Iter()
	var items []*pb.Span
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
		items = append(items, &span)
	}
	return items
}

func (cs *CassandraSource) GetSpanCount(ctx context.Context, traceID string) int64 {
	var count int64
	cs.CassandraSession.Session().Query("SELECT COUNT(trace_id) FROM spans WHERE trace_id = ?", traceID).Iter().Scan(&count)
	return count
}
