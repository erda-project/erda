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
	"context"
	"time"

	"github.com/olivere/elastic"

	"github.com/erda-project/erda/modules/core/monitor/metric"
	indexloader "github.com/erda-project/erda/modules/core/monitor/metric/index-loader"
)

// IndexMatcher .
type IndexMatcher func(index *indexloader.IndexEntry) bool

// Interface .
type Interface interface {
	GetWriteIndex(m *metric.Metric) (string, bool)
	GetWriteFixedIndex(m *metric.Metric) string
	CreateIndex(m *metric.Metric) error

	CleanIndices(ctx context.Context, filter IndexMatcher) error                                                               // 清理匹配且过期的索引数据
	RolloverIndices(ctx context.Context, filter IndexMatcher) error                                                            // 滚动匹配的索引
	MergeIndices(ctx context.Context, filter IndexMatcher, min string, merge, delete bool) ([]*MergeGroup, interface{}, error) // 合并匹配的索引

	IndexType() string
	IndexPrefix() string
	RequestTimeout() time.Duration
	Client() *elastic.Client
	EnableRollover() bool
}
