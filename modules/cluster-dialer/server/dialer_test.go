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
