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
	"net"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/discover"
)

var session TunnelSession

func init() {
	clusterDialerEndpoint := fmt.Sprintf("ws://%s%s", discover.ClusterDialer(), "/clusterdialer")
	go session.initialize(clusterDialerEndpoint)
}

type DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)
type DialContextProtoFunc func(ctx context.Context, address string) (net.Conn, error)

func DialContext(clusterKey string) DialContextFunc {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		logrus.Debugf("use cluster dialer, key:%s", clusterKey)
		f := session.getClusterDialer(ctx, clusterKey)
		if f == nil {
			return nil, errors.New("get cluster dialer failed")
		}
		return f(ctx, network, addr)
	}
}

func DialContextProto(clusterKey, proto string) DialContextProtoFunc {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		logrus.Debugf("use cluster dialer, key:%s", clusterKey)
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
