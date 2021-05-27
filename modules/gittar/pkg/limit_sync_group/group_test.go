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

import "testing"

func testLimitSyncGroup(t *testing.T, wg1 *limitSyncGroup, wg2 *limitSyncGroup) {
	n := 16
	wg1.Add(n)
	wg2.Add(n)
	exited := make(chan bool, n)
	for i := 0; i != n; i++ {
		go func() {
			wg1.Done()
			wg2.Wait()
			exited <- true
		}()
	}
	wg1.Wait()
	for i := 0; i != n; i++ {
		select {
		case <-exited:
			t.Fatal("WaitGroup released group too soon")
		default:
		}
		wg2.Done()
	}
	for i := 0; i != n; i++ {
		<-exited // Will block if barrier fails to unlock someone.
	}
}

func TestLimitSyncGroup(t *testing.T) {
	n := 5
	wg := NewSemaphore(n)
	wg.Add(n)
	go func() {
		for i := 0; i < n; i++ {
			wg.Done()
		}
	}()
	wg.Wait()
}

func TestLimitSyncGroup2(t *testing.T) {
	n := 20
	wg1 := NewSemaphore(n)
	wg2 := NewSemaphore(n)
	for i := 0; i != 20; i++ {
		testLimitSyncGroup(t, wg1, wg2)
	}
}
