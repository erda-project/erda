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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	clusteragent "github.com/erda-project/erda/modules/cluster/cluster-agent/client"
	clientconfig "github.com/erda-project/erda/modules/cluster/cluster-agent/config"
	"github.com/erda-project/erda/modules/cluster/cluster-dialer/auth"
	serverconfig "github.com/erda-project/erda/modules/cluster/cluster-dialer/config"
	"github.com/erda-project/erda/pkg/clusterdialer"
	"github.com/erda-project/erda/pkg/discover"
)

const (
	dialerListenAddr = "127.0.0.1:18751"
	helloListenAddr  = "127.0.0.1:18752"
)

func startServer(etcd *clientv3.Client) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go Start(ctx, nil, &serverconfig.Config{
		Listen:          dialerListenAddr,
		NeedClusterInfo: false,
	}, etcd)
	return ctx, cancel
}

type fakeKV struct {
	clientv3.KV
}

func (f *fakeKV) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return nil, nil
}

func Test_DialerContext(t *testing.T) {
	defer monkey.UnpatchAll()

	authorizer := auth.New(auth.WithCredentialClient(nil))
	monkey.Patch(authorizer.Authorizer, func(req *http.Request) (string, bool, error) {
		return fakeClusterKey, true, nil
	})

	client := clusteragent.New(clusteragent.WithConfig(&clientconfig.Config{
		ClusterDialEndpoint: fmt.Sprintf("ws://%s/clusteragent/connect", dialerListenAddr),
		ClusterKey:          fakeClusterKey,
		CollectClusterInfo:  false,
		ClusterAccessKey:    fakeClusterAccessKey,
	}))

	ctx, cancel := startServer(&clientv3.Client{KV: &fakeKV{}})
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
	}
	mx := mux.NewRouter()
	mx.HandleFunc("/hello", helloHandler)
	mx.HandleFunc("/clusterdialer/ip", queryIPFunc)
	go http.ListenAndServe(helloListenAddr, mx)

	go client.Start(ctx)
	for {
		if client.IsConnected() {
			logrus.Info("client connected")
			break
		}
		time.Sleep(1 * time.Second)
	}

	os.Setenv(discover.EnvClusterDialer, helloListenAddr)
	hc := http.Client{
		Transport: &http.Transport{
			DialContext: clusterdialer.DialContext("test"),
		},
		Timeout: 10 * time.Second,
	}
	req, _ := http.NewRequest("GET", "http://"+helloListenAddr, nil)
	_, err := hc.Do(req)
	if err != nil {
		t.Errorf("dialer failed, err:%+v", err)
	}
	cancel()
	select {
	case <-ctx.Done():
	}
}

func queryIPFunc(w http.ResponseWriter, req *http.Request) {
	res := map[string]interface{}{
		"succeeded": true,
		"IP":        dialerListenAddr,
	}
	data, _ := json.Marshal(res)
	io.WriteString(w, string(data))
}
