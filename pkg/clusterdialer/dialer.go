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
	"fmt"
	"net"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/discover"
)

var session TunnelSession
var initialized bool
var once sync.Once

func Init(addrOption ...string) {
	if initialized {
		return
	}
	var clusterDialAddr string
	if len(addrOption) == 0 {
		clusterDialAddr = discover.ClusterDialer()
		if clusterDialAddr == "" {
			return
		}
	} else {
		clusterDialAddr = addrOption[0]
	}
	clusterDialerEndpoint := fmt.Sprintf("ws://%s%s", clusterDialAddr, "/clusterdialer")
	go session.initialize(clusterDialerEndpoint)
	initialized = true
}

type DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)
type DialContextProtoFunc func(ctx context.Context, address string) (net.Conn, error)

func DialContext(clusterKey string) DialContextFunc {
	once.Do(func() { Init() })
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		logrus.Debugf("use cluster dialer, key:%s, addr:%s", clusterKey, addr)
		f := session.getClusterDialer(ctx, clusterKey)
		if f == nil {
			return nil, errors.New("get cluster dialer failed")
		}
		return f(ctx, network, addr)
	}
}

func DialContextProto(clusterKey, proto string) DialContextProtoFunc {
	once.Do(func() { Init() })
	return func(ctx context.Context, addr string) (net.Conn, error) {
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
