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

package goroutinepool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func blockF() {
	ch := make(chan struct{})
	<-ch
}

func TestBasic(t *testing.T) {
	p := New(5)
	p.Start()
	p.MustGo(blockF)
	p.MustGo(blockF)
	p.MustGo(blockF)
	p.MustGo(blockF)
	p.MustGo(blockF)
	err := p.Go(blockF)
	assert.NotNil(t, err)
}

func TestStartStop(t *testing.T) {
	r := 0
	p := New(5)
	p.Start()
	p.MustGo(func() {
		time.Sleep(1 * time.Second)
		r = 1
	})
	p.Stop()
	assert.Equal(t, r, 1)
}

func restart(p *GoroutinePool, r_ int) int {
	r := r_
	p.Start()
	p.MustGo(func() {
		time.Sleep(500 * time.Millisecond)
		r++
	})
	p.Stop()
	p.Start()
	p.MustGo(func() {
		time.Sleep(500 * time.Millisecond)
		r++
	})
	p.Stop()
	return r
}

//func TestRestart(t *testing.T) {
//	p := New(5)
//	r := restart(p, 0)
//	assert.Equal(t, r, 2)
//	r = restart(p, r)
//	assert.Equal(t, r, 4)
//	r = restart(p, r)
//	assert.Equal(t, r, 6)
//	r = restart(p, r)
//	assert.Equal(t, r, 8)
//}

func TestStat(t *testing.T) {
	p := New(10)
	stat := p.Statistics()
	assert.Equal(t, 0, stat[0])
	assert.Equal(t, 0, stat[1])
	p.Start()
	time.Sleep(500 * time.Millisecond)
	stat = p.Statistics()

	assert.Equal(t, 10, stat[0])
	assert.Equal(t, 10, stat[1])
	p.MustGo(func() {
		time.Sleep(1 * time.Second)
	})
	stat = p.Statistics()
	assert.Equal(t, 9, stat[0])
	assert.Equal(t, 10, stat[1])
	p.Stop()
}

//func TestJobPanic(t *testing.T) {
//	job := func() {
//		panic("panic")
//	}
//	p := New(1)
//	p.Start()
//	p.MustGo(job)
//	time.Sleep(500 * time.Millisecond)
//	stat := p.Statistics()
//	assert.Equal(t, 1, stat[0])
//	assert.Equal(t, 1, stat[1])
//	p.Stop()
//}
