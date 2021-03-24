package report

import "github.com/erda-project/erda/pkg/telemetry/common"

type buffer struct {
	count int
	max   int
	data  []*common.Metric
}

func newBuffer(max int) *buffer {
	b := new(buffer)

	b.data = make([]*common.Metric, max+1)
	b.count = 0
	b.max = max

	return b
}

func (b *buffer) IsOverFlow() bool {
	if b.count == b.max {
		return true
	}
	return false
}

func (b *buffer) IsEmpty() bool {
	if b.count == 0 {
		return true
	}
	return false
}

func (b *buffer) Flush() []*common.Metric {
	count := b.count
	b.count = 0
	if count > 0 {
		return b.data[:count]
	}

	return nil
}

func (b *buffer) Add(v *common.Metric) bool {
	if b.count < b.max {
		b.data[b.count] = v
		b.count++
		return true
	}

	return false
}
