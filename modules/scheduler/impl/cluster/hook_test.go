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

package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetDefaultClusterConfig(t *testing.T) {
	c1 := ClusterInfo{
		ClusterName: "x1",
		Kind:        "MARATHON",
		Options: map[string]string{
			"ADDR":      "http://dcos.x1.cn/service/marathon",
			"ENABLETAG": "true",
		},
	}

	setDefaultClusterConfig(&c1)
	assert.Equal(t, "true", c1.Options["ENABLETAG"])
	assert.Equal(t, "true", c1.Options["ENABLE_ORG"])
	assert.Equal(t, "true", c1.Options["ENABLE_WORKSPACE"])

	c2 := ClusterInfo{
		ClusterName: "x2",
		Kind:        "MARATHON",
		Options: map[string]string{
			"ADDR": "http://dcos.x2.cn/service/marathon",
		},
	}

	setDefaultClusterConfig(&c2)
	assert.Equal(t, "true", c2.Options["ENABLETAG"])
	assert.Equal(t, "true", c2.Options["ENABLE_ORG"])
	assert.Equal(t, "true", c2.Options["ENABLE_WORKSPACE"])

	c3 := ClusterInfo{
		ClusterName: "x3",
		Kind:        "METRONOME",
		Options: map[string]string{
			"ADDR": "http://dcos.x3.cn/service/metronome",
		},
	}

	setDefaultClusterConfig(&c3)
	assert.Equal(t, "true", c3.Options["ENABLETAG"])
	assert.Equal(t, "true", c3.Options["ENABLE_ORG"])
	assert.Equal(t, "true", c3.Options["ENABLE_WORKSPACE"])
}
