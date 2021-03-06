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

package enhancedqueue_test

//import (
//	"reflect"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/enhancedqueue"
//)
//
//func TestEnhancedQueue_Snapshot(t *testing.T) {
//	eq := enhancedqueue.NewEnhancedQueue(10)
//
//	// add new key to queue, priority: k2 > k3 > k1
//	now := time.Now().Round(0)
//	eq.Add("k1", 10, now)
//	eq.Add("k2", 20, now)
//	eq.Add("k3", 10, now.Add(-time.Second))
//
//	backup := eq.Export()
//
//	neq := enhancedqueue.NewEnhancedQueue(0)
//	err := neq.Import(backup)
//	assert.NoError(t, err)
//
//	assert.True(t, reflect.DeepEqual(eq.PendingQueue(), neq.PendingQueue()))
//	assert.True(t, reflect.DeepEqual(eq.ProcessingQueue(), neq.ProcessingQueue()))
//	assert.True(t, reflect.DeepEqual(eq.ProcessingWindow(), neq.ProcessingWindow()))
//	assert.True(t, reflect.DeepEqual(eq, neq))
//
//	assert.Equal(t, int64(10), eq.ProcessingWindow(), "processing window is 10")
//	assert.True(t, eq.InPending("k1"), "k1 in queue")
//	assert.True(t, eq.InPending("k2"), "k2 in queue")
//	assert.True(t, eq.InPending("k3"), "k3 in queue")
//	assert.Equal(t, "k2", eq.PopPending(true))
//	assert.Equal(t, "k2", eq.PopPending())
//	assert.Equal(t, "k3", eq.PopPending())
//	assert.Equal(t, "k1", eq.PopPending())
//	assert.Equal(t, "k1", eq.PopProcessing("k1", true))
//}
