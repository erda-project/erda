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

package indexmanager

import (
	"time"

	"github.com/olivere/elastic"

	"github.com/erda-project/erda/modules/core/monitor/metric"
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

	GetWriteIndex(m *metric.Metric) (string, bool)
	GetWriteFixedIndex(m *metric.Metric) string
	CreateIndex(m *metric.Metric) error

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
