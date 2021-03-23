package throttler_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/throttler"
)

func TestThrottler_Snapshot(t *testing.T) {
	th := throttler.NewNamedThrottler("default", nil)

	now := time.Now()
	th.AddQueue("q1", 10)
	th.AddQueue("q2", 1000000) // dangling queue
	th.AddKeyToQueues("k1", []throttler.AddKeyToQueueRequest{
		{
			QueueName:    "q1",
			QueueWindow:  &[]int64{20}[0],
			Priority:     100,
			CreationTime: now,
		},
		{
			QueueName:    "q3",
			Priority:     2,
			CreationTime: now.Add(-time.Hour * 3),
		},
	})

	backup := th.Export()

	nth := throttler.NewNamedThrottler("th", nil)
	err := nth.Import(backup)
	assert.NoError(t, err)

	deepEqual := reflect.DeepEqual(th, nth)
	assert.True(t, deepEqual)
}
