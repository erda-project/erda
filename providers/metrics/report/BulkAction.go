package report

type BulkAction struct {
	disruptor Disruptor
}

func (b *BulkAction) Push(request Metrics) error {
	return b.disruptor.In(request...)
}
