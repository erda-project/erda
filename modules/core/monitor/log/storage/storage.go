package storage

import (
	"context"

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
