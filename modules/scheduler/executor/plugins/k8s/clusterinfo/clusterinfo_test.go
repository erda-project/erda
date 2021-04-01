package clusterinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNetportalURL(t *testing.T) {
	var (
		testURL1 = "inet://127.0.0.1/master.mesos"
		testURL2 = "inet://127.0.0.2?ssl=on&direct=on/master.mesos/service/marathon?test=on"
	)

	netportal, err := parseNetportalURL(testURL1)
	assert.Nil(t, err)
	assert.Equal(t, "inet://127.0.0.1", netportal)

	netportal, err = parseNetportalURL(testURL2)
	assert.Nil(t, err)
	assert.Equal(t, "inet://127.0.0.2?ssl=on&direct=on", netportal)
}
