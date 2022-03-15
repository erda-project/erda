package source

import (
	"context"

	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
)

type TraceSource interface {
	GetSpans(ctx context.Context, req *pb.GetSpansRequest) []*pb.Span
	GetSpanCount(ctx context.Context, traceID string) int64
}
