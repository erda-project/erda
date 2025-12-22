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

package algo

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
)

// NextRoundRobinIndex returns the next index in [0,size) using store-backed counters.
func NextRoundRobinIndex(ctx context.Context, store state_store.LBStateStore, key state_store.CounterKey, size int) (int, error) {
	if size <= 0 {
		return -1, fmt.Errorf("size must be positive")
	}
	if store == nil {
		return -1, fmt.Errorf("nil state store")
	}
	counter, err := store.NextCounter(ctx, key)
	if err != nil {
		return -1, err
	}
	return int(counter % int64(size)), nil
}

// ConsistentHashIndex deterministically maps stickyValue to an index in [0,size).
// Returns -1 when size <= 0.
func ConsistentHashIndex(stickyValue string, size int) int {
	if size <= 0 {
		return -1
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(stickyValue))
	return int(h.Sum64() % uint64(size))
}
