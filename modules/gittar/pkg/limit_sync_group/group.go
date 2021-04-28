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

import "sync"

// encapsulate sync.WaitGroup
// realize a controllable number of WaitGroup
type limitSyncGroup struct {
	c  chan struct{}
	wg *sync.WaitGroup
}

func NewSemaphore(maxSize int) *limitSyncGroup {
	return &limitSyncGroup{
		c:  make(chan struct{}, maxSize),
		wg: new(sync.WaitGroup),
	}
}
func (s *limitSyncGroup) Add(delta int) {
	s.wg.Add(delta)
	for i := 0; i < delta; i++ {
		s.c <- struct{}{}
	}
}
func (s *limitSyncGroup) Done() {
	<-s.c
	s.wg.Done()
}
func (s *limitSyncGroup) Wait() {
	s.wg.Wait()
}
