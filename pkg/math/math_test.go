package math

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAbsInt(t *testing.T) {
	x := AbsInt(10)
	assert.Equal(t, 10, x)

	x = AbsInt(0)
	assert.Equal(t, 0, x)

	x = AbsInt(-10)
	assert.Equal(t, 10, x)
}

func TestAbsInt32(t *testing.T) {
	x := AbsInt32(10)
	assert.Equal(t, int32(10), x)

	x = AbsInt32(0)
	assert.Equal(t, int32(0), x)

	x = AbsInt32(-10)
	assert.Equal(t, int32(10), x)
}

func TestAbsInt64(t *testing.T) {
	x := AbsInt64(10)
	assert.Equal(t, int64(10), x)

	x = AbsInt64(0)
	assert.Equal(t, int64(0), x)

	x = AbsInt64(-10)
	assert.Equal(t, int64(10), x)
}
