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

package httpclientutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/stretchr/testify.v1/assert"
)

func TestRmProto(t *testing.T) {
	var expected = "dice.terminus.io"

	var http = "http://dice.terminus.io"
	require.Equal(t, RmProto(http), expected)

	var https = "https://dice.terminus.io"
	require.Equal(t, RmProto(https), expected)

	var noPrefix = "dice.terminus.io"
	require.Equal(t, RmProto(noPrefix), expected)
}

func TestGetPath(t *testing.T) {
	assert.Equal(t,
		"/repository/docker-hosted-platform/",
		GetPath("http://addon-nexus.default.svc.cluster.local:8081/repository/docker-hosted-platform/"))
}

func TestCombineHostAndPath(t *testing.T) {
	fmt.Println(CombineHostAndPath("https://baidu.com/", "/repository/platform/"))
}

func TestGetPort(t *testing.T) {
	//assert.Equal(t, 8081, GetPort("http://addon-nexus.default.svc.cluster.local:8081/repository/docker-hosted-platform/"))
	//assert.Equal(t, 80, GetPort("addon-nexus"))
	assert.Equal(t, 8081, GetPort("addon-nexus.default.svc.cluster.local:8081"))
}

func TestGetHost(t *testing.T) {
	assert.Equal(t, "nexus-sys.dev.terminus.io", GetHost("http://nexus-sys.dev.terminus.io/repository/platform/"))
}
