// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package uintset

import (
	"sync"
)

type Uint64Set struct {
	mu           sync.RWMutex
	seriesIDSet  map[uint64]struct{}
	seriesIDList []uint64
}

func NewUint64Set(initCap int) *Uint64Set {
	return &Uint64Set{seriesIDSet: make(map[uint64]struct{}, initCap), seriesIDList: make([]uint64, 0, initCap)}
}

func (ss *Uint64Set) Has(x uint64) bool {
	ss.mu.RLock()
	_, ok := ss.seriesIDSet[x]
	ss.mu.RUnlock()
	if ok {
		return true
	}
	return false
}

func (ss *Uint64Set) Add(x uint64) {
	ss.mu.Lock()
	ss.seriesIDSet[x] = struct{}{}
	ss.seriesIDList = append(ss.seriesIDList, x)
	ss.mu.Unlock()
}

func (ss *Uint64Set) AddBatch(batch []uint64) {
	ss.mu.Lock()
	for _, x := range batch {
		ss.seriesIDSet[x] = struct{}{}
		ss.seriesIDList = append(ss.seriesIDList, x)
	}
	ss.mu.Unlock()
}

func (ss *Uint64Set) CleanOldPart() {
	ss.mu.Lock()
	n := len(ss.seriesIDList) / 2
	toDelete := ss.seriesIDList[:n]
	ss.seriesIDList = ss.seriesIDList[n:]
	for i := range toDelete {
		delete(ss.seriesIDSet, toDelete[i])
	}
	ss.mu.Unlock()
}
