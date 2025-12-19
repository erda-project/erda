package state_store

import (
	"context"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	redis "github.com/go-redis/redis"
)

func TestRedisStateStore(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	prefix := "test-lb"
	store := NewRedisStateStore(client, prefix)
	ctx := context.Background()

	// empty binding
	if _, ok, err := store.GetBinding(ctx, "bk", "sv"); err != nil || ok {
		t.Fatalf("expected empty binding, ok=%v err=%v", ok, err)
	}

	if err := store.SetBinding(ctx, "bk", "sv", "ins-1", 500*time.Millisecond); err != nil {
		t.Fatalf("set binding failed: %v", err)
	}
	if val, ok, err := store.GetBinding(ctx, "bk", "sv"); err != nil || !ok || val != "ins-1" {
		t.Fatalf("unexpected binding result val=%s ok=%v err=%v", val, ok, err)
	}
	keys := mr.Keys()
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %v", keys)
	}
	expectPrefix := prefix + ":branch-bind:bk:sticky:"
	if !strings.HasPrefix(keys[0], expectPrefix) {
		t.Fatalf("unexpected binding key format: %s", keys[0])
	}

	first, err := store.NextCounter(ctx, "counter/1")
	if err != nil || first != 1 {
		t.Fatalf("unexpected first counter: %d err=%v", first, err)
	}
	second, err := store.NextCounter(ctx, "counter/1")
	if err != nil || second != 2 {
		t.Fatalf("unexpected second counter: %d err=%v", second, err)
	}

	mr.FastForward(time.Second)
	if _, ok, err := store.GetBinding(ctx, "bk", "sv"); err != nil || ok {
		t.Fatalf("expected binding expired, ok=%v err=%v", ok, err)
	}
}

func TestRedisStateStoreUniversal(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	opt := &redis.UniversalOptions{Addrs: []string{mr.Addr()}, DB: 0}
	prefix := "test-lb"
	store := NewRedisStateStoreUniversal(opt, prefix)
	if store == nil {
		t.Fatalf("expected universal store not nil")
	}
	ctx := context.Background()
	if err := store.SetBinding(ctx, "bk", "sv", "ins-1", time.Second); err != nil {
		t.Fatalf("set binding failed: %v", err)
	}
	if val, ok, err := store.GetBinding(ctx, "bk", "sv"); err != nil || !ok || val != "ins-1" {
		t.Fatalf("unexpected binding result val=%s ok=%v err=%v", val, ok, err)
	}
	found := false
	for _, k := range mr.Keys() {
		if strings.HasPrefix(k, prefix+":branch-bind:bk:sticky:") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected branch-bind key in redis")
	}
}
