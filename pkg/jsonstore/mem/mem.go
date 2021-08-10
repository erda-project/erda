// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package mem

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
)

type MemStore struct {
	pool sync.Map
}

func New() (*MemStore, error) {
	store := &MemStore{
		pool: sync.Map{},
	}
	return store, nil
}

func (s *MemStore) Put(ctx context.Context, key, value string) error {
	s.pool.Store(key, value)
	return nil
}

func (s *MemStore) PutWithOption(ctx context.Context, key, value string, opts []interface{}) (interface{}, error) {
	return nil, s.Put(ctx, key, value)
}

func (s *MemStore) Get(ctx context.Context, key string) (storetypes.KeyValue, error) {
	value, ok := s.pool.Load(key)
	if !ok {
		return storetypes.KeyValue{}, errors.New("not found")
	}
	return storetypes.KeyValue{Key: []byte(key), Value: []byte(value.(string))}, nil
}

// TODO: better impl, instead of O(n)?
func (s *MemStore) PrefixGet(ctx context.Context, prefix string) ([]storetypes.KeyValue, error) {
	r := []storetypes.KeyValue{}
	s.pool.Range(func(k, v interface{}) bool {
		k_ := k.(string)
		if strings.HasPrefix(k_, prefix) {
			r = append(r, storetypes.KeyValue{Key: []byte(k_), Value: []byte(v.(string))})
		}
		return true
	})
	return r, nil
}

func (s *MemStore) Remove(ctx context.Context, key string) (*storetypes.KeyValue, error) {
	value, ok := s.pool.Load(key)
	if !ok {
		return nil, nil
	}
	s.pool.Delete(key)
	return &storetypes.KeyValue{
		Key:   []byte(key),
		Value: []byte(value.(string)),
	}, nil

}

func (s *MemStore) PrefixRemove(ctx context.Context, prefix string) ([]storetypes.KeyValue, error) {
	r := []storetypes.KeyValue{}
	s.pool.Range(func(k, v interface{}) bool {
		k_ := k.(string)
		if strings.HasPrefix(k_, prefix) {
			s.pool.Delete(k)
			r = append(r, storetypes.KeyValue{Key: []byte(k_), Value: []byte(v.(string))})
		}
		return true
	})
	return r, nil

}
func (s *MemStore) PrefixGetKey(ctx context.Context, prefix string) ([]storetypes.Key, error) {
	r := []storetypes.Key{}
	s.pool.Range(func(k, v interface{}) bool {
		k_ := k.(string)
		if strings.HasPrefix(k_, prefix) {
			r = append(r, storetypes.Key([]byte(k_)))
		}
		return true
	})
	return r, nil
}
