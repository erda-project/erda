package state_store

import (
	"context"
	"time"
)

type (
	// BindingKey identifies a sticky binding namespace (e.g. per client/group).
	BindingKey = string
	// CounterKey identifies a counter namespace (e.g. per branch).
	CounterKey = string
)

// LBStateStore abstracts sticky binding and counters for load balancing.
type LBStateStore interface {
	GetBinding(ctx context.Context, bindingKey BindingKey, stickyValue string) (instanceID string, ok bool, err error)
	SetBinding(ctx context.Context, bindingKey BindingKey, stickyValue, instanceID string, ttl time.Duration) error
	NextCounter(ctx context.Context, key CounterKey) (int64, error)
}
