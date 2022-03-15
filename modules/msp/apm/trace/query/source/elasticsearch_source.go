package source

import (
	"context"

	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
	"github.com/erda-project/erda/modules/msp/apm/trace"
	"github.com/erda-project/erda/modules/msp/apm/trace/storage"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
)

type ElasticsearchSource struct {
	StorageReader storage.Storage
}

func (esc ElasticsearchSource) GetSpans(ctx context.Context, req *pb.GetSpansRequest) []*pb.Span {
	org := req.OrgName
	if len(org) <= 0 {
		org = apis.GetHeader(ctx, "org")
	}
	// do esc query
	elasticsearchSpans, _ := fetchSpanFromES(ctx, esc.StorageReader, storage.Selector{
		TraceId: req.TraceID,
		Hint: storage.QueryHint{
			Scope:     org,
			Timestamp: req.StartTime * 1000000, // convert ms to ns
		},
	}, true, int(req.GetLimit()))
	var spans []*pb.Span
	for _, value := range elasticsearchSpans {
		var span pb.Span
		span.Id = value.SpanId
		span.TraceId = value.TraceId
		span.OperationName = value.OperationName
		span.ParentSpanId = value.ParentSpanId
		span.StartTime = value.StartTime
		span.EndTime = value.EndTime
		span.Tags = value.Tags
		spans = append(spans, &span)
	}
	return spans
}

func (esc *ElasticsearchSource) GetSpanCount(ctx context.Context, traceID string) int64 {
	return esc.StorageReader.Count(ctx, traceID)
}

func fetchSpanFromES(ctx context.Context, storage storage.Storage, sel storage.Selector, forward bool, limit int) (list []*trace.Span, err error) {
	it, err := storage.Iterator(ctx, &sel)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	defer it.Close()

	if forward {
		for it.Next() {
			span, ok := it.Value().(*trace.Span)
			if !ok {
				continue
			}
			list = append(list, span)
			if len(list) >= limit {
				break
			}
		}
	} else {
		for it.Prev() {
			span, ok := it.Value().(*trace.Span)
			if !ok {
				continue
			}
			list = append(list, span)
			if len(list) >= limit {
				break
			}
		}
	}
	if it.Error() != nil {
		return nil, errors.NewInternalServerError(err)
	}

	return list, it.Error()
}
