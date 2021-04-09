package trace

// Span .
type Span struct {
	TraceID       string            `json:"trace_id"`
	StartTime     int64             `json:"start_time"`
	SpanID        string            `json:"span_id"`
	ParentSpanID  string            `json:"parent_span_id"`
	OperationName string            `json:"operation_name"`
	EndTime       int64             `json:"end_time"`
	Tags          map[string]string `json:"tags"`
}
