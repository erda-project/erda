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

package clusterdialer

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/discover"
)

const HandshakeTimeOut = 10 * time.Second

var sessionG *remotedialer.Session
var lock sync.Mutex

func init() {
	headers := http.Header{
		"X-API-Tunnel-Proxy": []string{"on"},
	}
	dialer := &websocket.Dialer{
		HandshakeTimeout: HandshakeTimeOut,
	}
	go func() {
		for {
			clusterDialerUrl := fmt.Sprintf("ws://%s%s", discover.ClusterDialer(), "/clusterdialer")
			ws, _, err := dialer.Dial(clusterDialerUrl, headers)
			if err != nil {
				logrus.Errorf("Failed to connect to proxy server %s, err: %v", clusterDialerUrl, err)
				time.Sleep(time.Duration(rand.Int()%10) * time.Second)
				continue
			}
			lock.Lock()
			sessionG = remotedialer.NewClientSession(func(string, string) bool { return true }, ws)
			lock.Unlock()
			_, err = sessionG.Serve(context.Background())
			if err != nil {
				logrus.Errorf("Failed to serve proxy connection err: %v", err)
			}
			lock.Lock()
			sessionG = nil
			lock.Unlock()
			sessionG.Close()
			ws.Close()
			// retry connect after sleep a random time
			time.Sleep(time.Duration(rand.Int()%10) * time.Second)
		}
	}()
}

func getClusterDialer(ctx context.Context, clusterKey string) remotedialer.Dialer {
	var session *remotedialer.Session
	start := time.Now()
	for {
		lock.Lock()
		session = sessionG
		lock.Unlock()
		if session != nil {
			break
		}
		select {
		case <-ctx.Done():
			logrus.Errorf("get clusterdial session failed, cost %.3fs", time.Since(start).Seconds())
			return nil
		case <-time.After(1 * time.Second):
			logrus.Infof("waiting fo clusterdial session ready... ")
		}
	}
	return remotedialer.ToDialer(session, clusterKey)
}

type DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)
type DialContextProtoFunc func(ctx context.Context, address string) (net.Conn, error)

func DialContext(clusterKey string) DialContextFunc {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		logrus.Debug("use cluster dialer, key:%s", clusterKey)
		return getClusterDialer(ctx, clusterKey)(ctx, network, addr)
	}
}

func DialContextProto(clusterKey, proto string) DialContextProtoFunc {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		logrus.Debug("use cluster dialer, key:%s", clusterKey)
		return getClusterDialer(ctx, clusterKey)(ctx, proto, addr)
	}
}

func DialContextTCP(clusterKey string) DialContextProtoFunc {
	return DialContextProto(clusterKey, "tcp")
}
