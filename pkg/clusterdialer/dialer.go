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

package clusterdialer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/cache"
	"github.com/erda-project/erda/pkg/discover"
)

var (
	sessions = sync.Map{}
	ipCache  = cache.New("clusterDialerEndpoint", time.Second*30, queryClusterDialerIP)
)

type DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)
type DialContextProtoFunc func(ctx context.Context, address string) (net.Conn, error)

func DialContext(clusterKey string) DialContextFunc {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		session, err := getSession(clusterKey)
		if err != nil {
			return nil, errors.Errorf("failed to get session for clusterKey %s, %v", clusterKey, err)
		}
		logrus.Debugf("use cluster dialer, key:%s, addr:%s", clusterKey, addr)
		f := session.getClusterDialer(ctx, clusterKey)
		if f == nil {
			return nil, errors.New("get cluster dialer failed")
		}
		return f(ctx, network, addr)
	}
}

func DialContextProto(clusterKey, proto string) DialContextProtoFunc {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		session, err := getSession(clusterKey)
		if err != nil {
			return nil, errors.Errorf("failed to get session for clusterKey %s, %v", clusterKey, err)
		}
		logrus.Debugf("use cluster dialer, key:%s, addr:%s", clusterKey, addr)
		f := session.getClusterDialer(ctx, clusterKey)
		if f == nil {
			return nil, errors.New("get cluster dialer failed")
		}
		return f(ctx, proto, addr)
	}
}

func DialContextTCP(clusterKey string) DialContextProtoFunc {
	return DialContextProto(clusterKey, "tcp")
}

func queryClusterDialerIP(clusterKey interface{}) (interface{}, bool) {
	logrus.Debugf("start querying clusterDialer IP...")
	host := "http://" + discover.ClusterDialer()
	resp, err := http.Get(host + fmt.Sprintf("/clusterdialer/ip?clusterKey=%s", clusterKey))
	if err != nil {
		logrus.Errorf("failed to request clsuterdialer in cache updating, %v", err)
		return "", false
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("failed to read from resp in cache updating, %v", err)
		return "", false
	}
	r := make(map[string]interface{})
	if err = json.Unmarshal(data, &r); err != nil {
		logrus.Errorf("failed to unmarshal resp in cache updating, %v", err)
		return "", false
	}

	succeeded, _ := r["succeeded"].(bool)
	if !succeeded {
		errStr, _ := r["error"].(string)
		logrus.Errorf("return error from clusterdialer in cache updating, %s", errStr)
		return "", false
	}

	ip, _ := r["IP"].(string)
	return ip, true
}

func getSession(clusterKey string) (*TunnelSession, error) {
	clusterDialerEndpoint, ok := ipCache.LoadWithUpdate(clusterKey)
	if !ok {
		logrus.Errorf("failed to get clusterDialer endpoint for clusterKey %s", clusterKey)
		return nil, errors.Errorf("failed to get clusterDialer endpoint for clusterKey %s", clusterKey)
	}
	if clusterDialerEndpoint == "" {
		return nil, errors.Errorf("can not found clusterDialer for clusterKey %s", clusterKey)
	}
	logrus.Debugf("[DEBUG] get cluster dialer endpoint succeeded, IP: %s", clusterDialerEndpoint)

	var session *TunnelSession
	v, ok := sessions.Load(clusterDialerEndpoint)
	if !ok {
		ctx, cancel := context.WithCancel(context.Background())
		session = &TunnelSession{expired: ctx, cancel: cancel, clusterDialerEndpoint: clusterDialerEndpoint.(string)}
		go session.initialize(fmt.Sprintf("ws://%s%s", clusterDialerEndpoint, "/clusterdialer"))
		sessions.Store(clusterDialerEndpoint, session)
	} else {
		session, _ = v.(*TunnelSession)
	}
	return session, nil
}
