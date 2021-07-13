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

package mysqlhelper

import (
	"net"
	"testing"
)

func TestSplitHost(t *testing.T) {
	host, port, err := net.SplitHostPort("mysql-slave.group-addon-mysql--df161e4cd23934dfeb75b96203a7970ea.svc.cluster.local:3306")
	if err != nil {
		t.Errorf("should not error")
	}
	if port != "3306" {
		t.Error("port error")
	}
	if host != "mysql-slave.group-addon-mysql--df161e4cd23934dfeb75b96203a7970ea.svc.cluster.local" {
		t.Errorf("host error")
	}
}
