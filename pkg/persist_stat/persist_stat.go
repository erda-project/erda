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

package persist_stat

import (
	"sync"
	"time"
)

type AccumType int

const (
	SUM AccumType = iota
	AVG
)

type Policy struct {
	AccumTp      AccumType
	Interval     int // second
	PreserveDays int
}

type PersistStoreStat interface {

	// 对于etcd后端，interval >= 60s，以控制数据量 1day <= 60 * 24 (1440)
	SetPolicy(policy Policy) error

	Emit(tag string, value int64) error

	Last5Min() (map[string]int64, error)

	Last20Min() (map[string]int64, error)

	Last1Hour() (map[string]int64, error)

	Last6Hour() (map[string]int64, error)

	Last1Day() (map[string]int64, error)

	Stat(beginTimestamp, endTimestamp time.Time) (map[string]int64, error)

	Clear(beforeTimeStamp time.Time) error

	// Metrics 返回自启动以来的 metric 统计
	Metrics() map[string]int64
}

type MemMetrics struct {
	sync.RWMutex
	m map[string]int64
}

func NewMemMetrics() *MemMetrics {
	m := map[string]int64{}
	return &MemMetrics{m: m}
}

func (m *MemMetrics) EmitMetric(tag string, value int64) {
	m.Lock()
	defer m.Unlock()
	v, ok := m.m[tag]
	if !ok {
		m.m[tag] = value
		return
	}
	m.m[tag] = v + value
}
func (m *MemMetrics) GetMetrics() map[string]int64 {
	m.RLock()
	defer m.RUnlock()
	newm := map[string]int64{}
	for tag, v := range m.m {
		newm[tag] = v
	}
	return newm
}
