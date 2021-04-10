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

package util

type GPool struct {
	queue chan int
}

func NewGPool(size int) *GPool {
	if size <= 0 {
		size = 1
	}
	return &GPool{
		queue: make(chan int, size),
	}
}

func (p *GPool) Acquire(delta int) {
	for i := 0; i < delta; i++ {
		p.queue <- 1
	}
}

func (p *GPool) Release(delta int) {
	for i := 0; i < delta; i++ {
		<-p.queue
	}
}
