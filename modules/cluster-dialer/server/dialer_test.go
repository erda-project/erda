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
	"testing"
	"time"

	"github.com/erda-project/erda/modules/cluster-agent/client"
	clientconfig "github.com/erda-project/erda/modules/cluster-agent/config"
	serverconfig "github.com/erda-project/erda/modules/cluster-dialer/config"
	"github.com/erda-project/erda/pkg/clusterdialer"
)

const (
	dialerListenAddr = "127.0.0.1:18751"
	helloListenAddr  = "127.0.0.1:18752"
)

func init() {
	clusterdialer.Init(dialerListenAddr)
}

func startServer() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go Start(ctx, &serverconfig.Config{
		Listen:          dialerListenAddr,
		NeedClusterInfo: false,
	})
	return ctx, cancel
}

func Test_DialerContext(t *testing.T) {
	ctx, cancel := startServer()
	go client.Start(context.Background(), &clientconfig.Config{
		ClusterDialEndpoint: fmt.Sprintf("ws://%s/clusteragent/connect", dialerListenAddr),
		ClusterKey:          "test",
		SecretKey:           "test",
		CollectClusterInfo:  false,
	})

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
	}
	http.HandleFunc("/hello", helloHandler)
	go http.ListenAndServe(helloListenAddr, nil)
	select {
	case <-client.Connected():
		time.Sleep(1 * time.Second)
		fmt.Println("client connected")
	}
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
