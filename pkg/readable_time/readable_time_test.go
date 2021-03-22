package readable_time

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadableTime(t *testing.T) {
	t1, err := time.Parse(time.RFC3339, "2018-12-05T16:54:57+08:00")
	assert.Nil(t, err)
	t2, err := time.Parse(time.RFC3339, "2018-12-05T16:54:59+08:00")
	assert.Nil(t, err)

	a := readableTime(t1, t2)
	assert.Equal(t, int64(2), a.Second)
}
