package state_store

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	redis "github.com/go-redis/redis"
)

// RedisStateStore implements LBStateStore backed by Redis.
// It is safe for multi-instance deployments when all instances share the same Redis.
type RedisStateStore struct {
	client redis.Cmdable
	prefix string
}

// NewRedisStateStore builds a Redis-backed LBStateStore.
// prefix is optional and defaults to "ai-proxy:lb" if empty.
func NewRedisStateStore(client redis.Cmdable, prefix string) *RedisStateStore {
	if client == nil {
		return nil
	}
	if prefix == "" {
		prefix = "ai-proxy:lb"
	}
	return &RedisStateStore{client: client, prefix: prefix}
}

// NewRedisStateStoreUniversal builds store using redis.UniversalOptions (supports standalone, sentinel, or cluster).
// Caller owns the lifecycle of the created client.
func NewRedisStateStoreUniversal(opt *redis.UniversalOptions, prefix string) *RedisStateStore {
	if opt == nil {
		return nil
	}
	client := redis.NewUniversalClient(opt)
	return NewRedisStateStore(client, prefix)
}

func (s *RedisStateStore) GetBinding(_ context.Context, bindingKey BindingKey, stickyValue string) (string, bool, error) {
	key := s.bindingKey(bindingKey, stickyValue)
	val, err := s.client.Get(key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (s *RedisStateStore) SetBinding(_ context.Context, bindingKey BindingKey, stickyValue, instanceID string, ttl time.Duration) error {
	key := s.bindingKey(bindingKey, stickyValue)
	if ttl <= 0 {
		ttl = time.Hour
	}
	return s.client.Set(key, instanceID, ttl).Err()
}

func (s *RedisStateStore) NextCounter(_ context.Context, key CounterKey) (int64, error) {
	return s.client.Incr(s.counterKey(key)).Result()
}

func (s *RedisStateStore) bindingKey(bindingKey BindingKey, stickyValue string) string {
	hash := hashSticky(stickyValue)
	return fmt.Sprintf("%s:branch-bind:%s:sticky:%s", s.prefix, bindingKey, hash)
}

func (s *RedisStateStore) counterKey(key CounterKey) string {
	return fmt.Sprintf("%s:counter:%s", s.prefix, key)
}

func hashSticky(v string) string {
	h := sha1.Sum([]byte(v))
	return hex.EncodeToString(h[:])[:16]
}
