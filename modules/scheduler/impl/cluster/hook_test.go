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
