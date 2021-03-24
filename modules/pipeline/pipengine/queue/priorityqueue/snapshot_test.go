package priorityqueue_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/priorityqueue"
)

func TestPriorityQueue_Snapshot(t *testing.T) {
	pq := priorityqueue.NewPriorityQueue()

	now := time.Now().Round(0)
	pq.Add(priorityqueue.NewItem("k1", 10, now))
	pq.Add(priorityqueue.NewItem("k2", 20, now.Add(time.Second)))
	backup := pq.Export()

	npq := priorityqueue.NewPriorityQueue()
	err := npq.Import(backup)
	assert.NoError(t, err, "import from export data")
	assert.Equal(t, 2, npq.Len())
	assert.Equal(t, "k2", npq.Peek().Key())
	assert.Equal(t, int64(10), npq.Get("k1").Priority())
	assert.Equal(t, int64(20), npq.Get("k2").Priority())
}
