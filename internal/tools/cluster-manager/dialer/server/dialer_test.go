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

package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/rancher/remotedialer"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/cluster-manager/dialer/auth"
	"github.com/erda-project/erda/internal/tools/cluster-manager/dialer/config"
	"github.com/erda-project/erda/pkg/clusterdialer"
	"github.com/erda-project/erda/pkg/discover"
)

const (
	dialerListenAddr = "127.0.0.1"
	dialerListenPort = "18751"
)

func startServer(etcd *clientv3.Client) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go Start(ctx, &fakeClusterSvc{}, nil, &config.Config{
		Listen:          fmt.Sprintf("%s:%s", dialerListenAddr, dialerListenPort),
		NeedClusterInfo: false,
		AuthWhitelist:   auth.SkipAllKey,
	}, etcd, &bundle.Bundle{})
	return ctx, cancel
}

type fakeKV struct {
	clientv3.KV
}

func (f *fakeKV) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return &clientv3.GetResponse{
		Kvs: []*mvccpb.KeyValue{{
			Value: []byte(dialerListenAddr),
		}},
	}, nil
}

func (f *fakeKV) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return nil, nil
}

type fakeClusterSvc struct {
	clusterpb.ClusterServiceServer
}

func (f *fakeClusterSvc) GetCluster(context.Context, *clusterpb.GetClusterRequest) (*clusterpb.GetClusterResponse, error) {
	return &clusterpb.GetClusterResponse{Data: &clusterpb.ClusterInfo{Name: fakeClusterKey}}, nil
}

func (f *fakeClusterSvc) PatchCluster(context.Context, *clusterpb.PatchClusterRequest) (*clusterpb.PatchClusterResponse, error) {
	return &clusterpb.PatchClusterResponse{}, nil
}

func Test_DialerContext(t *testing.T) {
	ctx, cancel := startServer(&clientv3.Client{KV: &fakeKV{}})
	_ = cancel
	waitForDialerServer(t, fmt.Sprintf("%s:%s", dialerListenAddr, dialerListenPort), fakeClusterKey)
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!")
	}
	mx := mux.NewRouter()
	mx.HandleFunc("/hello", helloHandler)
	helloServer := httptest.NewServer(mx)
	defer helloServer.Close()

	startTestClusterAgent(t, ctx, fmt.Sprintf("%s:%s", dialerListenAddr, dialerListenPort), fakeClusterKey, fakeClusterAccessKey)
	waitForClusterSession(t, fmt.Sprintf("%s:%s", dialerListenAddr, dialerListenPort), fakeClusterKey)

	os.Setenv(discover.EnvClusterDialer, fmt.Sprintf("%s:%s", dialerListenAddr, dialerListenPort))
	hc := http.Client{
		Transport: &http.Transport{
			DialContext: clusterdialer.DialContext(fakeClusterKey),
		},
		Timeout: 10 * time.Second,
	}
	req, _ := http.NewRequest("GET", helloServer.URL+"/hello", nil)
	resp, err := hc.Do(req)
	if err != nil {
		t.Fatalf("dialer failed, err:%+v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status:%d expect:200", resp.StatusCode)
		return
	}
	respBody, _ := io.ReadAll(resp.Body)
	if string(respBody) != "Hello, world!" {
		t.Errorf("respBody:%s, expect:Hello, world!", respBody)
		return
	}
}

func startTestClusterAgent(t *testing.T, ctx context.Context, dialerAddr, clusterKey, accessKey string) {
	t.Helper()

	headers := http.Header{
		apistructs.ClusterManagerHeaderKeyClusterKey.String():    {clusterKey},
		apistructs.ClusterManagerHeaderKeyAuthorization.String(): {accessKey},
	}
	endpoint := fmt.Sprintf("ws://%s/clusteragent/connect", dialerAddr)

	go func() {
		for {
			err := remotedialer.ConnectToProxy(ctx, endpoint, headers, func(proto, address string) bool {
				return proto == "tcp"
			}, nil, func(ctx context.Context, _ *remotedialer.Session) error {
				<-ctx.Done()
				return nil
			})
			if err == nil || ctx.Err() != nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
}
