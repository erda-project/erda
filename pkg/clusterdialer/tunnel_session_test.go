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
	"testing"
	"time"
)

func TestTunnelSession_getClusterDialer_timeout(t *testing.T) {
	var session TunnelSession
	go session.initialize("ws://127.0.0.1")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	dialer := session.getClusterDialer(ctx, "test")
	if dialer != nil {
		t.Error("dialer is not nil")
	}
}
