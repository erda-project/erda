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
