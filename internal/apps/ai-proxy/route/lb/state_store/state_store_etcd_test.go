package state_store

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

func TestEtcdStateStore(t *testing.T) {
	cfg := embed.NewConfig()
	cfg.Logger = "zap"
	cfg.Dir = t.TempDir()

	lc := mustFreeURL(t)
	lp := mustFreeURL(t)
	cfg.ListenClientUrls = []url.URL{lc}
	cfg.AdvertiseClientUrls = cfg.ListenClientUrls
	cfg.ListenPeerUrls = []url.URL{lp}
	cfg.AdvertisePeerUrls = cfg.ListenPeerUrls
	cfg.InitialCluster = fmt.Sprintf("%s=%s", cfg.Name, cfg.AdvertisePeerUrls[0].String())

	etcd, err := embed.StartEtcd(cfg)
	if err != nil {
		t.Fatalf("failed to start embed etcd: %v", err)
	}
	defer etcd.Close()

	select {
	case <-etcd.Server.ReadyNotify():
	case <-time.After(10 * time.Second):
		t.Fatalf("embedded etcd start timeout")
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{cfg.ListenClientUrls[0].String()},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("failed to create etcd client: %v", err)
	}
	defer cli.Close()

	store := NewEtcdStateStore(cli, "test-lb")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// empty binding
	if _, ok, err := store.GetBinding(ctx, "bk", "sv"); err != nil || ok {
		t.Fatalf("expected empty binding, ok=%v err=%v", ok, err)
	}

	if err := store.SetBinding(ctx, "bk", "sv", "ins-1", time.Second); err != nil {
		t.Fatalf("set binding failed: %v", err)
	}
	if val, ok, err := store.GetBinding(ctx, "bk", "sv"); err != nil || !ok || val != "ins-1" {
		t.Fatalf("unexpected binding val=%s ok=%v err=%v", val, ok, err)
	}

	first, err := store.NextCounter(ctx, "counter/1")
	if err != nil || first != 1 {
		t.Fatalf("unexpected first counter %d err=%v", first, err)
	}
	second, err := store.NextCounter(ctx, "counter/1")
	if err != nil || second != 2 {
		t.Fatalf("unexpected second counter %d err=%v", second, err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		_, ok, err := store.GetBinding(ctx, "bk", "sv")
		if err != nil {
			t.Fatalf("get binding failed: %v", err)
		}
		if !ok {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected binding expired before %v", deadline)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func mustFreeURL(t *testing.T) url.URL {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen for free port: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	parsed, err := url.Parse("http://" + addr)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}
	return *parsed
}
