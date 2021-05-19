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
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"
)

const HandshakeTimeOut = 10 * time.Second

type TunnelSession struct {
	session *remotedialer.Session
	lock    sync.Mutex
}

func (s *TunnelSession) initialize(endpoint string) {
	headers := http.Header{
		"X-API-Tunnel-Proxy": []string{"on"},
	}
	dialer := &websocket.Dialer{
		HandshakeTimeout: HandshakeTimeOut,
	}
	for {
		ws, _, err := dialer.Dial(endpoint, headers)
		if err != nil {
			logrus.Errorf("Failed to connect to proxy server %s, err: %v", endpoint, err)
			time.Sleep(time.Duration(rand.Int()%10) * time.Second)
			continue
		}
		s.lock.Lock()
		s.session = remotedialer.NewClientSession(func(string, string) bool { return true }, ws)
		s.lock.Unlock()
		_, err = s.session.Serve(context.Background())
		if err != nil {
			logrus.Errorf("Failed to serve proxy connection err: %v", err)
		}
		s.session.Close()
		s.lock.Lock()
		s.session = nil
		s.lock.Unlock()
		ws.Close()
		// retry connect after sleep a random time
		time.Sleep(time.Duration(rand.Int()%10) * time.Second)
	}

}

func (s *TunnelSession) getClusterDialer(ctx context.Context, clusterKey string) remotedialer.Dialer {
	var session *remotedialer.Session
	start := time.Now()
	for {
		s.lock.Lock()
		session = s.session
		s.lock.Unlock()
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
