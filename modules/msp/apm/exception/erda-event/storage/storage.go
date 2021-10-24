package storage

import "github.com/erda-project/erda/modules/core/monitor/storekit"
import "context"

type (
	// Selector .
	Selector struct {
		StartTime int64
		EndTime int64
		EventId string
		ErrorId string
		TerminusKey string
	}

	// Storage .
	Storage interface {
		NewWriter(ctx context.Context) (storekit.BatchWriter, error)
		Iterator(ctx context.Context, sel *Selector) (storekit.Iterator, error)
		Count(ctx context.Context, sel *Selector) int64
	}
)
