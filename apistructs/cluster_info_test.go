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
