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
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/cluster-manager/dialer/auth"
	"github.com/erda-project/erda/internal/tools/cluster-manager/dialer/config"
)

const (
	dialerListenAddr2    = "127.0.0.1:18753"
	fakeClusterKey       = "test"
	fakeClusterAccessKey = "init"
)

func Test_netportal(t *testing.T) {
	serverCtx := context.Background()
	go Start(serverCtx, &fakeClusterSvc{}, nil, &config.Config{
		Listen:          dialerListenAddr2,
		NeedClusterInfo: false,
		AuthWhitelist:   auth.SkipAllKey,
	}, &clientv3.Client{KV: &fakeKV{}}, bundle.New())

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!")
	}
	mx := mux.NewRouter()
	mx.HandleFunc("/hello2", helloHandler)
	helloServer := httptest.NewServer(mx)
	defer helloServer.Close()
	helloURL, err := url.Parse(helloServer.URL)
	if err != nil {
		t.Fatalf("parse hello server url: %v", err)
	}

	tctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := serverHealthCheck(tctx); err != nil {
		t.Fatal(err)
	}

	clientCtx := context.Background()
	startTestClusterAgent(t, clientCtx, dialerListenAddr2, fakeClusterKey, fakeClusterAccessKey)
	waitForClusterSession(t, dialerListenAddr2, fakeClusterKey)
	hc := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+dialerListenAddr2+"/hello2", nil)
	req.Header = http.Header{
		portalHostHeader:    {fakeClusterKey},
		portalDestHeader:    {helloURL.Host},
		portalTimeoutHeader: {"10"},
	}
	req.URL.RawQuery = "query=ut"
	resp, err := hc.Do(req)
	if err != nil {
		t.Errorf("request failed, err:%+v", err)
		return
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

func serverHealthCheck(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("server health check timeout")
		default:
			hc := &http.Client{}
			resp, err := hc.Get(fmt.Sprintf("http://%s/clusterdialer/ip?clusterKey=%s",
				dialerListenAddr2, fakeClusterKey))
			if err != nil {
				logrus.Errorf("server health check failed, err:%+v", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
}

func waitForClusterSession(t *testing.T, dialerAddr, clusterKey string) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("http://%s/clusteragent/check?clusterKey=%s", dialerAddr, clusterKey))
		if err == nil {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr == nil && resp.StatusCode == http.StatusOK && string(body) == "true" {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("cluster session not ready for %s", clusterKey)
}

func waitForDialerServer(t *testing.T, dialerAddr, clusterKey string) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("http://%s/clusterdialer/ip?clusterKey=%s", dialerAddr, clusterKey))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("dialer server not ready for %s", dialerAddr)
}
