package storage

import (
	"context"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
)

type (
	// Selector .
	Selector struct {
		TraceId string
	}

	// Storage .
	Storage interface {
		NewWriter(ctx context.Context) (storekit.BatchWriter, error)
		Iterator(ctx context.Context, sel *Selector) (storekit.Iterator, error)
		Count(ctx context.Context, traceId string) int64
	}

)
