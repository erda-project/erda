package loop

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoop_CalculateInterval(t *testing.T) {
	// 2, 4, 8, 10, 10
	l := New(WithMaxTimes(math.MaxUint64), WithDeclineLimit(10*time.Second), WithDeclineRatio(2))

	interval := l.CalculateInterval(0)
	assert.Equal(t, time.Second*2, interval)

	interval = l.CalculateInterval(1)
	assert.Equal(t, time.Second*4, interval)

	interval = l.CalculateInterval(2)
	assert.Equal(t, time.Second*8, interval)

	interval = l.CalculateInterval(3)
	assert.Equal(t, time.Second*10, interval)

	interval = l.CalculateInterval(4)
	assert.Equal(t, time.Second*10, interval)
}
