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

package cache

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/go-redis/redis"

	"github.com/erda-project/erda-proto-go/msp/apm/checker/pb"
)

type (
	Key   = int64
	Value = *pb.Checker
)

// Cache .
type Cache struct {
	root  string
	redis *redis.Client
}

// New .
func New(root string, redis *redis.Client) *Cache {
	return &Cache{
		root:  root,
		redis: redis,
	}
}

func (c *Cache) Keys() ([]Key, error) {
	list, err := c.redis.HKeys(c.root).Result()
	if err != nil {
		return nil, err
	}
	var keys []Key
	for _, item := range list {
		v, err := strconv.ParseInt(item, 10, 64)
		if err != nil {
			continue
		}
		keys = append(keys, v)
	}
	return keys, nil
}

func (c *Cache) ListByKeys(keys []Key) (map[Key]Value, error) {
	if len(keys) <= 0 {
		return nil, nil
	}
	var fields []string
	for _, k := range keys {
		fields = append(fields, strconv.FormatInt(k, 10))
	}
	vals, err := c.redis.HMGet(c.root, fields...).Result()
	if err != nil {
		return nil, err
	}
	values := make(map[Key]Value)
	for i, val := range vals {
		if v, ok := val.(string); ok && len(v) > 0 {
			var out pb.Checker
			err := json.NewDecoder(strings.NewReader(v)).Decode(&out)
			if err != nil {
				continue
			}
			values[keys[i]] = &out
		}
	}
	return values, nil
}

func (c *Cache) Put(data *pb.Checker) error {
	if data == nil {
		return nil
	}
	sb := &strings.Builder{}
	_ = json.NewEncoder(sb).Encode(data)
	return c.redis.HSet(c.root, strconv.FormatInt(data.Id, 10), sb.String()).Err()
}

func (c *Cache) Remove(key Key) error {
	return c.redis.HDel(c.root, strconv.FormatInt(key, 10)).Err()
}

func (c *Cache) Contains(key Key) (bool, error) {
	return c.redis.HExists(c.root, strconv.FormatInt(key, 10)).Result()
}

// CacheStorage .
type CacheStorage struct {
	*Cache
}

func (s *CacheStorage) ListIDs() ([]int64, error) {
	return s.Cache.Keys()
}

func (s *CacheStorage) ListByIDs(ids []int64) (map[int64]*pb.Checker, error) {
	return s.Cache.ListByKeys(ids)
}
