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

package odata

import (
	"sync"
)

// Buffer stores ObservableData in a circular buffer.
// https://en.wikipedia.org/wiki/Circular_buffer
type Buffer struct {
	sync.Mutex
	buf        []ObservableData
	start, end int
	cap, size  int
}

func NewBuffer(cap int) *Buffer {
	return &Buffer{
		buf:   make([]ObservableData, cap),
		start: 0,
		end:   0,
		cap:   cap,
		size:  0,
	}
}

func (b *Buffer) Push(od ObservableData) {
	b.Lock()
	defer b.Unlock()
	if b.size == b.cap {
		return
	}
	b.buf[b.end] = od

	b.end = b.next(b.end)
	b.size++
}

func (b *Buffer) Pop() ObservableData {
	b.Lock()
	defer b.Unlock()
	if b.size == 0 {
		return nil
	}
	item := b.buf[b.start]

	b.start = b.next(b.start)
	b.size--
	return item
}

func (b *Buffer) Full() bool {
	b.Lock()
	defer b.Unlock()
	return b.size == b.cap
}

func (b *Buffer) Empty() bool {
	return b.size == 0
}

// FlushAll empty buffer, then return a copy of internal buf
func (b *Buffer) FlushAll() []ObservableData {
	b.Lock()
	defer b.Unlock()
	out := make([]ObservableData, b.size)
	for i := range out {
		out[i] = b.buf[b.start]
		b.buf[b.start] = nil
		b.start = b.next(b.start)
	}
	b.size = 0
	return out
}

func (b *Buffer) next(index int) int {
	index++
	if index == b.cap {
		return 0
	}
	return index
}

func (b *Buffer) nextBy(index, offset int) int {
	index += offset
	index %= b.cap
	return index
}
