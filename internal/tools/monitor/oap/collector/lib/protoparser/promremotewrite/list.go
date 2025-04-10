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

package promremotewrite

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/recallsong/go-utils/encoding/md5x"
)

func MapHash(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for key, _ := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	var sb = strings.Builder{}
	for _, key := range keys {
		sb.WriteString(fmt.Sprintf("%v:%v\n", key, m[key]))
	}
	return md5x.SumString(sb.String()).String()
}

type GroupMetricBucket map[MetricTagHash]*metric.Metric
type GroupMetricList struct {
	data []*GroupMetricListItem
	lock sync.RWMutex
}

func NewGroupMetricList() *GroupMetricList {
	return &GroupMetricList{
		data: make([]*GroupMetricListItem, 0),
		lock: sync.RWMutex{},
	}
}

type GroupMetricListItem struct {
	bucket    GroupMetricBucket
	timestamp int64
}

func (list *GroupMetricList) Length() int {
	list.lock.Lock()
	defer list.lock.Unlock()
	length := len(list.data)
	return length
}

// Append /
// append metric into bucket
func (list *GroupMetricList) Append(m *metric.Metric) {
	list.lock.Lock()
	defer list.lock.Unlock()

	for index := range list.data {
		if list.data[index].timestamp == m.Timestamp {
			list.data[index].bucket.Set(m)
			return
		}

		if list.data[index].timestamp > m.Timestamp {
			list.data = append(list.data[:index+1], list.data[index:]...)
			list.data[index] = &GroupMetricListItem{
				bucket:    make(GroupMetricBucket),
				timestamp: m.Timestamp,
			}
			list.data[index].bucket.Set(m)
			return
		}
	}

	list.data = append(list.data, &GroupMetricListItem{
		bucket:    make(GroupMetricBucket),
		timestamp: m.Timestamp,
	})
	list.data[len(list.data)-1].bucket.Set(m)
}

// ListMetrics /**
// list all metrics
func (list *GroupMetricList) ListMetrics() []*metric.Metric {
	list.lock.Lock()
	defer list.lock.Unlock()
	var res = list.listMetrics()
	return res
}

func (list *GroupMetricList) listMetrics() []*metric.Metric {
	var res = make([]*metric.Metric, 0)
	for index := range list.data {
		for _, m := range list.data[index].bucket {
			res = append(res, m)
		}
	}
	return res
}

// PopMetrics /***
// pop metrics in first bucket
func (list *GroupMetricList) PopMetrics() []*metric.Metric {
	list.lock.Lock()
	defer list.lock.Unlock()
	if len(list.data) == 0 {
		return nil
	}
	var res = make([]*metric.Metric, 0)
	for _, m := range (list.data)[0].bucket {
		res = append(res, m)
	}
	list.data = list.data[1:]
	return res
}

func (list *GroupMetricList) PopAllMetrics() []*metric.Metric {
	list.lock.Lock()
	defer list.lock.Unlock()
	metrics := list.listMetrics()
	//list.data = make([]*GroupMetricListItem, 0)
	return metrics
}

// Set /
// set metric into bucket
func (hashMap *GroupMetricBucket) Set(m *metric.Metric) {
	tagHash := MapHash(m.Tags)
	mc, ok := (*hashMap)[MetricTagHash(tagHash)]
	if !ok {
		mc = &metric.Metric{
			Name:      m.Name,
			Timestamp: m.Timestamp,
			Tags:      m.Tags,
			Fields:    make(map[string]interface{}),
			OrgName:   m.OrgName,
		}
	}
	for k, v := range m.Fields {
		mc.Fields[k] = v
	}
	(*hashMap)[MetricTagHash(tagHash)] = mc
}
