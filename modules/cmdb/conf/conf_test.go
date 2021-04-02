package conf

import (
	"strings"
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestGetComponentName(t *testing.T) {
	assert.Equal(t, "addon-nexus", strings.Split("addon-nexus.default:8081", ".")[0])
}
