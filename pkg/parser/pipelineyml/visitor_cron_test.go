package pipelineyml

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

const (
	everyMin = "0 * * * ? *"
)

func TestListNextCronTime(t *testing.T) {
	a, err := ListNextCronTime("0 */10 * * * *")
	fmt.Println(a)
	assert.NoError(t, err)

	b, err := ListNextCronTime("*/10 * * * *")
	fmt.Println(b)
	assert.NoError(t, err)

	fmt.Println(reflect.DeepEqual(a, b))
}

func TestParseCronExpr(t *testing.T) {
	a, err := ListNextCronTime("0 * * * * ? *")
	assert.NoError(t, err)
	spew.Dump(a)
}

func TestListNextCronTime2(t *testing.T) {
	var (
		nextTimes []time.Time
		err       error
	)
	// specify nothing
	nextTimes, err = ListNextCronTime(everyMin)
	assert.NoError(t, err)
	assert.True(t, len(nextTimes) == defaultListNextScheduleCount)

	// specify count=2
	nextTimes, err = ListNextCronTime(everyMin, WithListNextScheduleCount(2))
	assert.NoError(t, err)
	assert.True(t, len(nextTimes) == 2)

	// specify count=0
	nextTimes, err = ListNextCronTime(everyMin, WithListNextScheduleCount(0))
	assert.NoError(t, err)
	assert.True(t, len(nextTimes) == 0)

	// specify count=-1
	nextTimes, err = ListNextCronTime(everyMin, WithListNextScheduleCount(-1))
	assert.NoError(t, err)
	assert.True(t, len(nextTimes) == maxListNextScheduleCount)
	fmt.Println(nextTimes[maxListNextScheduleCount-1])

	// specify cronStartTime
	cronStartTime := time.Date(2020, 3, 1, 1, 2, 3, 0, time.UTC)
	nextTimes, err = ListNextCronTime(everyMin, WithCronStartEndTime(&cronStartTime, nil))
	assert.NoError(t, err)
	assert.True(t, len(nextTimes) == defaultListNextScheduleCount)
	fmt.Println(nextTimes[0])

	// specify cronEndTime
	alreadyCronEndTime := time.Date(1999, 3, 1, 1, 2, 3, 0, time.UTC)
	nextTimes, err = ListNextCronTime(everyMin, WithCronStartEndTime(nil, &alreadyCronEndTime))
	assert.NoError(t, err)
	fmt.Println(nextTimes)
	assert.True(t, len(nextTimes) == 0)

	// specify cronStartTime and cronEndTime
	nextTimes, err = ListNextCronTime(everyMin, WithCronStartEndTime(&cronStartTime, &alreadyCronEndTime))
	assert.NoError(t, err)
	assert.True(t, len(nextTimes) == 0)

	// specify cronStartTime and cronEndTime and count
	cronEndTime := cronStartTime.Add(time.Minute)
	nextTimes, err = ListNextCronTime(everyMin, WithCronStartEndTime(&cronStartTime, &cronEndTime), WithListNextScheduleCount(-2))
	assert.NoError(t, err)
	assert.True(t, len(nextTimes) == 1)

	// specify cronStartTime and cronEndTime and count
	cronEndTime = cronStartTime.Add(time.Minute * 10).Add(-time.Second * 4)
	nextTimes, err = ListNextCronTime(everyMin, WithCronStartEndTime(&cronStartTime, &cronEndTime), WithListNextScheduleCount(-1))
	assert.NoError(t, err)
	assert.True(t, len(nextTimes) == 9)
}
