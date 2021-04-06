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

package apistructs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var clusterInfo ClusterInfoData

func init() {
	clusterInfo = ClusterInfoData{
		DICE_PROTOCOL:    "http,https",
		DICE_HTTP_PORT:   "80",
		DICE_HTTPS_PORT:  "443",
		DICE_ROOT_DOMAIN: "dev.terminus.io",
	}
}

func TestClusterInfoData_DiceProtocolIsHTTPS(t *testing.T) {
	clusterInfo := ClusterInfoData{
		DICE_PROTOCOL: "http,https",
	}
	assert.True(t, clusterInfo.DiceProtocolIsHTTPS())
	clusterInfo = ClusterInfoData{
		DICE_PROTOCOL: "http",
	}
	assert.False(t, clusterInfo.DiceProtocolIsHTTPS())
}

func TestClusterInfoData_MustGetPublicURL(t *testing.T) {
	assert.Equal(t, "https://soldier.dev.terminus.io:443", clusterInfo.MustGetPublicURL("soldier"))
}
