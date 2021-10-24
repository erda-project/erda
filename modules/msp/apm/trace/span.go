package trace

type Span struct {
	TraceId       string            `json:"trace_id"`
	SpanId        string            `json:"span_id"`
	ParentSpanId  string            `json:"parent_span_id"`
	OperationName string            `json:"operation_name"`
	StartTime     int64             `json:"start_time"`
	EndTime       int64             `json:"end_time"`
	Tags          map[string]string `json:"tags"`
}
