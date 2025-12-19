package state_store

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdStateStore implements LBStateStore backed by etcd.
type EtcdStateStore struct {
	client *clientv3.Client
	prefix string
}

// NewEtcdStateStore builds an etcd-backed LBStateStore.
// prefix is optional and defaults to "ai-proxy:lb" if empty.
func NewEtcdStateStore(client *clientv3.Client, prefix string) *EtcdStateStore {
	if client == nil {
		return nil
	}
	if prefix == "" {
		prefix = "ai-proxy:lb"
	}
	return &EtcdStateStore{client: client, prefix: prefix}
}

func (s *EtcdStateStore) GetBinding(ctx context.Context, bindingKey BindingKey, stickyValue string) (string, bool, error) {
	key := s.bindingKey(bindingKey, stickyValue)
	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return "", false, err
	}
	if len(resp.Kvs) == 0 {
		return "", false, nil
	}
	return string(resp.Kvs[0].Value), true, nil
}

func (s *EtcdStateStore) SetBinding(ctx context.Context, bindingKey BindingKey, stickyValue, instanceID string, ttl time.Duration) error {
	key := s.bindingKey(bindingKey, stickyValue)
	leaseTTL := int64(math.Ceil(ttl.Seconds()))
	if leaseTTL <= 0 {
		leaseTTL = int64((time.Hour).Seconds())
	}
	lease, err := s.client.Grant(ctx, leaseTTL)
	if err != nil {
		return err
	}
	_, err = s.client.Put(ctx, key, instanceID, clientv3.WithLease(lease.ID))
	return err
}

func (s *EtcdStateStore) NextCounter(ctx context.Context, key CounterKey) (int64, error) {
	k := s.counterKey(key)
	for {
		resp, err := s.client.Get(ctx, k)
		if err != nil {
			return 0, err
		}
		var current int64
		var version int64
		if len(resp.Kvs) > 0 {
			current, _ = strconv.ParseInt(string(resp.Kvs[0].Value), 10, 64)
			version = resp.Kvs[0].Version
		}
		next := current + 1
		txn := s.client.Txn(ctx).If(clientv3.Compare(clientv3.Version(k), "=", version)).
			Then(clientv3.OpPut(k, strconv.FormatInt(next, 10))).
			Else()
		txnResp, err := txn.Commit()
		if err != nil {
			return 0, err
		}
		if txnResp.Succeeded {
			return next, nil
		}
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func (s *EtcdStateStore) bindingKey(bindingKey BindingKey, stickyValue string) string {
	return fmt.Sprintf("%s:binding:%s:%s", s.prefix, bindingKey, stickyValue)
}

func (s *EtcdStateStore) counterKey(key CounterKey) string {
	return fmt.Sprintf("%s:counter:%s", s.prefix, key)
}
