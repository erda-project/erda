package uuid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUuid(t *testing.T) {
	assert.Equal(t, 32, len(Generate()))
}
