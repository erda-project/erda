// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
