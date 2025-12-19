// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
