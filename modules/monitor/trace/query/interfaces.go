package query

type (
	SpanQueryAPI interface {
		SelectSpans(traceId string, limit int64) []map[string]interface{}
	}
)

func (p *provider) SelectSpans(traceId string, limit int64) []map[string]interface{} {
	return p.selectSpans(traceId, limit)
}
