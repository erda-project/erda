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

package cachetypes

import (
	"context"
)

// CacheItem defines the interface for individual cache items
type CacheItem interface {
	// ListAll returns all cached items, fallback to DB if cache miss
	ListAll(ctx context.Context) (uint64, any, error)
	// GetByID returns an item by ID, fallback to DB if cache miss
	GetByID(ctx context.Context, id string) (any, error)
	// Refresh refreshes the cache from database
	Refresh() (uint64, error)
}

// Manager defines the interface for cache manager
type Manager interface {
	// ListAll returns all cached items of the specified type
	ListAll(ctx context.Context, itemType ItemType) (uint64, any, error)
	// GetByID returns an item by ID for the specified type
	GetByID(ctx context.Context, itemType ItemType, id string) (any, error)
}
