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
	"sync"
	"time"
)

// MemoryStateStore is an in-memory LBStateStore for single-instance deployments.
type MemoryStateStore struct {
	mu       sync.Mutex
	counters map[string]int64
	bindings map[string]map[string]bindingEntry // bindingKey -> stickyValue -> entry
	now      func() time.Time
}

type bindingEntry struct {
	instanceID string
	expireAt   time.Time
}

func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		counters: make(map[string]int64),
		bindings: make(map[string]map[string]bindingEntry),
		now:      time.Now,
	}
}

func (s *MemoryStateStore) GetBinding(_ context.Context, bindingKey BindingKey, stickyValue string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	group, ok := s.bindings[bindingKey]
	if !ok {
		return "", false, nil
	}
	entry, ok := group[stickyValue]
	if !ok {
		return "", false, nil
	}
	if s.now().After(entry.expireAt) {
		delete(group, stickyValue)
		if len(group) == 0 {
			delete(s.bindings, bindingKey)
		}
		return "", false, nil
	}
	return entry.instanceID, true, nil
}

func (s *MemoryStateStore) SetBinding(_ context.Context, bindingKey BindingKey, stickyValue, instanceID string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.bindings[bindingKey] == nil {
		s.bindings[bindingKey] = make(map[string]bindingEntry)
	}
	s.bindings[bindingKey][stickyValue] = bindingEntry{
		instanceID: instanceID,
		expireAt:   s.now().Add(ttl),
	}
	return nil
}

func (s *MemoryStateStore) NextCounter(_ context.Context, key CounterKey) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.counters[key] + 1
	s.counters[key] = next
	return next, nil
}
