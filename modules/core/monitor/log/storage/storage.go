package storage

import (
	"context"
	"math"

	"github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
)

type (
	// Operator .
	Operator int32
	// Filter .
	Filter struct {
		Key   string
		Op    Operator
		Value interface{}
	}

	// Selector .
	Selector struct {
		Start   int64
		End     int64
		Scheme  string
		Filters []*Filter
		Debug   bool
		Options map[string]interface{}
	}

	// Storage .
	Storage interface {
		NewWriter(ctx context.Context) (storekit.BatchWriter, error)
		Iterator(ctx context.Context, sel *Selector) (storekit.Iterator, error)
	}
)

const (
	// EQ equal
	EQ Operator = iota
	REGEXP
)

// Comparer .
type Comparer struct{}

// DefaultComparer
var DefaultComparer = Comparer{}

var _ storekit.Comparer = (*Comparer)(nil)

func (c Comparer) Compare(a, b interface{}) int {
	al, ok := a.(*pb.LogItem)
	if !ok {
		return -1
	}
	bl, ok := b.(*pb.LogItem)
	if !ok {
		return -1
	}
	if al.Timestamp > bl.Timestamp {
		return 1
	} else if al.Timestamp < bl.Timestamp {
		return -1
	}
	if al.Offset != math.MaxInt64 && bl.Offset != math.MaxInt64 {
		if al.Offset > bl.Offset {
			return 1
		} else if al.Offset < bl.Offset {
			return -1
		}
	} else if al.Content != bl.Content {
		if al.Offset > bl.Offset {
			return 1
		} else if al.Offset < bl.Offset {
			return -1
		}
		if al.Content > bl.Content {
			return 1
		}
		return -1
	}
	return 0
}
