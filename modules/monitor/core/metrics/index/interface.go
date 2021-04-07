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

package indexmanager

import (
	"time"

	"github.com/erda-project/erda/modules/monitor/core/metrics"
	"github.com/olivere/elastic"
)

// IndexEntry .
type IndexEntry struct {
	Index       string
	Metric      string
	Namespace   string
	Key         string
	Fixed       bool
	Num         int64
	MinT        time.Time
	MaxT        time.Time
	DocsCount   int
	DocsDeleted int
	StoreSize   int64
	Active      bool
	// Deleted     int32
}

// IndexMatcher .
type IndexMatcher func(index *IndexEntry) bool

// Index .
type Index interface {
	WaitIndicesLoad()

	GetReadIndices(metrics []string, namespace []string, start, end int64) []string

	GetWriteIndex(m *metrics.Metric) (string, bool)
	GetWriteFixedIndex(m *metrics.Metric) string
	CreateIndex(m *metrics.Metric) error

	CleanIndices(filter IndexMatcher) error                                                               // 清理匹配且过期的索引数据
	RolloverIndices(filter IndexMatcher) error                                                            // 滚动匹配的索引
	MergeIndices(filter IndexMatcher, min string, merge, delete bool) ([]*MergeGroup, interface{}, error) // 合并匹配的索引

	MetricNames() []string
	EmptyIndex() string
	IndexPrefix() string
	IndexType() string
	RequestTimeout() time.Duration
	Client() *elastic.Client
	URLs() string
	EnableRollover() bool
}

// IndexEntrys .
type IndexEntrys []*IndexEntry

func (entrys IndexEntrys) Len() int      { return len(entrys) }
func (entrys IndexEntrys) Swap(i, j int) { entrys[i], entrys[j] = entrys[j], entrys[i] }
func (entrys IndexEntrys) Less(i, j int) bool {
	if entrys[i].Num == entrys[j].Num {
		if entrys[i].MinT.Equal(entrys[j].MinT) {
			if entrys[i].MaxT.Equal(entrys[j].MaxT) {
				return entrys[i].StoreSize < entrys[j].StoreSize
			}
			return entrys[i].MaxT.Before(entrys[j].MaxT)
		}
		return entrys[i].MinT.Before(entrys[j].MinT)
	}
	return entrys[i].Num < entrys[j].Num
}
