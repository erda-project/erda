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
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/erda-project/erda/modules/cluster-agent/client"
	clientconfig "github.com/erda-project/erda/modules/cluster-agent/config"
	serverconfig "github.com/erda-project/erda/modules/cluster-dialer/config"
)

const (
	dialerListenAddr = "127.0.0.1:18753"
	helloListenAddr  = "127.0.0.1:18754"
)

func Test_netportal(t *testing.T) {
	go Start(context.Background(), &serverconfig.Config{
		Listen:          dialerListenAddr,
		NeedClusterInfo: false,
	})
	go client.Start(context.Background(), &clientconfig.Config{
		ClusterDialEndpoint: fmt.Sprintf("ws://%s/clusteragent/connect", dialerListenAddr),
		ClusterKey:          "test",
		SecretKey:           "test",
		CollectClusterInfo:  false,
	})
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!")
	}
	http.HandleFunc("/hello", helloHandler)
	go http.ListenAndServe(helloListenAddr, nil)
	select {
	case <-client.Connected():
		fmt.Println("client connected")
	}
	hc := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+dialerListenAddr+"/hello", nil)
	req.Header = http.Header{
		portalSchemeHeader:  {"http"},
		portalHostHeader:    {"test"},
		portalDestHeader:    {helloListenAddr},
		portalTimeoutHeader: {"10"},
	}
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
