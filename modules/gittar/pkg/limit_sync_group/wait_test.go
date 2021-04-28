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

package limit_sync_group

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_limitSyncGroup(t *testing.T) {
	var table = []struct {
		num           int
		concurrentNum int
	}{
		{
			num:           3,
			concurrentNum: 10,
		},
		{
			num:           4,
			concurrentNum: 5,
		},
		{
			num:           11,
			concurrentNum: 2,
		},
		{
			num:           2,
			concurrentNum: 2,
		},
		{
			num:           11,
			concurrentNum: 10,
		},
	}
	for _, data := range table {
		var preGoroutineNum = runtime.NumGoroutine()
		var wait = NewSemaphore(data.concurrentNum)
		for i := 0; i < data.num; i++ {
			wait.Add(1)
			go func() {
				defer wait.Done()
				diffGoroutineNum := runtime.NumGoroutine() - preGoroutineNum
				assert.Less(t, diffGoroutineNum, data.concurrentNum+1)
			}()
		}
		wait.Wait()
	}
}
