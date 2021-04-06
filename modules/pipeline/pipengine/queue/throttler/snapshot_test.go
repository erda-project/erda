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

package throttler_test

//import (
//	"reflect"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/throttler"
//)
//
//func TestThrottler_Snapshot(t *testing.T) {
//	th := throttler.NewNamedThrottler("default", nil)
//
//	now := time.Now()
//	th.AddQueue("q1", 10)
//	th.AddQueue("q2", 1000000) // dangling queue
//	th.AddKeyToQueues("k1", []throttler.AddKeyToQueueRequest{
//		{
//			QueueName:    "q1",
//			QueueWindow:  &[]int64{20}[0],
//			Priority:     100,
//			CreationTime: now,
//		},
//		{
//			QueueName:    "q3",
//			Priority:     2,
//			CreationTime: now.Add(-time.Hour * 3),
//		},
//	})
//
//	backup := th.Export()
//
//	nth := throttler.NewNamedThrottler("th", nil)
//	err := nth.Import(backup)
//	assert.NoError(t, err)
//
//	deepEqual := reflect.DeepEqual(th, nth)
//	assert.True(t, deepEqual)
//}
