package loop

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoopMaxTimes(t *testing.T) {
	count := 0
	l := New(WithMaxTimes(20))
	l.Do(func() (bool, error) {
		count += 1
		return false, nil
	})
	assert.Equal(t, 20, count)
}

func TestLoopInterval(t *testing.T) {
	l := New(WithMaxTimes(10), WithInterval(time.Second*2))

	start := time.Now()
	l.Do(func() (bool, error) {
		return false, nil
	})
	duration := time.Now().Sub(start)
	if duration < 20*time.Second {
		t.Error("interval is not good working")
	}
}

func TestLoopRaiot(t *testing.T) {
	l := New(WithDeclineRatio(1.5), WithDeclineLimit(time.Second*10))
	begin := time.Now()
	l.Do(func() (bool, error) {
		executeTime := time.Now()
		fmt.Println("interval:", executeTime.Sub(begin))
		begin = executeTime
		return false, errors.New("effect")
	})
}
