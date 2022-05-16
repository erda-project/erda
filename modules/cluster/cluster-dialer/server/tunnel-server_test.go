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
	"io/ioutil"
	"net/http"
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
)

const (
	dialerListenAddr2    = "127.0.0.1:18753"
	helloListenAddr2     = "127.0.0.1:18754"
	fakeClusterKey       = "test"
	fakeClusterAccessKey = "init"
)

func Test_netportal(t *testing.T) {
	defer monkey.UnpatchAll()

	authorizer := auth.New(auth.WithCredentialClient(nil))
	monkey.Patch(authorizer.Authorizer, func(req *http.Request) (string, bool, error) {
		return fakeClusterKey, true, nil
	})

	client := clusteragent.New(clusteragent.WithConfig(&clientconfig.Config{
		ClusterDialEndpoint: fmt.Sprintf("ws://%s/clusteragent/connect", dialerListenAddr2),
		ClusterKey:          fakeClusterKey,
		CollectClusterInfo:  false,
		ClusterAccessKey:    fakeClusterAccessKey,
	}))

	go Start(context.Background(), &fakeClusterSvc{}, nil, &serverconfig.Config{
		Listen:          dialerListenAddr2,
		NeedClusterInfo: false,
	}, &clientv3.Client{KV: &fakeKV{}})
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!")
	}
	mx := mux.NewRouter()
	mx.HandleFunc("/hello2", helloHandler)
	mx.HandleFunc("/clusterdialer/ip", queryIPFunc2)
	go http.ListenAndServe(helloListenAddr2, mx)

	go client.Start(context.Background())
	for {
		if client.IsConnected() {
			logrus.Info("client connected")
			break
		}
		time.Sleep(1 * time.Second)
	}
	hc := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+dialerListenAddr2+"/hello2", nil)
	req.Header = http.Header{
		portalHostHeader:    {"test"},
		portalDestHeader:    {helloListenAddr2},
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
	respBody, _ := ioutil.ReadAll(resp.Body)
	if string(respBody) != "Hello, world!" {
		t.Errorf("respBody:%s, expect:Hello, world!", respBody)
		return
	}
}

func queryIPFunc2(w http.ResponseWriter, req *http.Request) {
	res := map[string]interface{}{
		"succeeded": true,
		"IP":        dialerListenAddr2,
	}
	data, _ := json.Marshal(res)
	io.WriteString(w, string(data))
}
