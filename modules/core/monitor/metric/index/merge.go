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
	"fmt"
	"sort"
	"strings"

	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/encoding/jsonx"
	"github.com/recallsong/go-utils/encoding/md5x"
	"github.com/recallsong/go-utils/lang/size"
)

// MergeGroup .
type MergeGroup struct {
	List        []*IndexEntry
	indices     []string
	Indices     string
	MergedSize  string
	MergedDocs  int
	MergedIndex string
}

// MergeIndices .
func (m *IndexManager) MergeIndices(filter IndexMatcher, min string, merge, delete bool) ([]*MergeGroup, interface{}, error) {
	minSize, err := size.ParseBytes(min)
	if err != nil {
		return nil, nil, err
	}
	v := m.indices.Load()
	if v == nil {
		return nil, nil, nil
	}
	indices := v.(map[string]*indexGroup)
	var merges []*MergeGroup
	for metric, mg := range indices {
		for ns, ng := range mg.Groups {
			var list []*IndexEntry
			var curSize int64
			for i := len(ng.List) - 1; i >= 0; i-- {
				entry := ng.List[i]
				if filter(entry) && m.needToMerge(entry, minSize-curSize) {
					list = append(list, entry)
					curSize += entry.StoreSize
				} else {
					if len(list) > 0 {
						if len(list) > 1 {
							mg := &MergeGroup{
								List:       list,
								MergedSize: size.FormatBytes(curSize),
							}
							m.initMergeGroup(metric, ns, "", mg)
							merges = append(merges, mg)
						}
						list, curSize = nil, 0
					}
				}
			}
			if len(list) > 0 {
				if len(list) > 1 {
					mg := &MergeGroup{
						List:       list,
						MergedSize: size.FormatBytes(curSize),
					}
					m.initMergeGroup(metric, ns, "", mg)
					merges = append(merges, mg)
				}
				list, curSize = nil, 0
			}
			for key, kg := range ng.Groups {
				for i := len(kg.List) - 1; i >= 0; i-- {
					entry := kg.List[i]
					if filter(entry) && m.needToMerge(entry, minSize-curSize) {
						list = append(list, entry)
						curSize += entry.StoreSize
					} else {
						if len(list) > 0 {
							if len(list) > 1 {
								mg := &MergeGroup{
									List:       list,
									MergedSize: size.FormatBytes(curSize),
								}
								m.initMergeGroup(metric, ns, key, mg)
								merges = append(merges, mg)
							}
							list, curSize = nil, 0
						}
					}
				}
				if len(list) > 0 {
					if len(list) > 1 {
						mg := &MergeGroup{
							List:       list,
							MergedSize: size.FormatBytes(curSize),
						}
						m.initMergeGroup(metric, ns, key, mg)
					}
					list, curSize = nil, 0
				}
			}
		}
	}
	var resps []interface{}
	if merge {
		for i, merge := range merges {
			resp, err := m.doIndicesMerge(merge, delete)
			resps = append(resps, resp)
			if err != nil {
				if i > 0 {
					m.toReloadIndices(false)
				}
				return merges, resps, err
			}
		}
		if len(merges) > 0 {
			m.toReloadIndices(false)
		}
	}
	return merges, resps, nil
}

func (m *IndexManager) needToMerge(index *IndexEntry, size int64) bool {
	if !index.Fixed && !index.Active && index.StoreSize >= 0 && index.StoreSize < size {
		return true
	}
	return false
}

func (m *IndexManager) initMergeGroup(metric, ns, key string, merge *MergeGroup) {
	var indices []string
	var docs int
	for _, item := range merge.List {
		indices = append(indices, item.Index)
		docs += item.DocsCount
	}
	merge.MergedDocs = docs
	sort.Strings(indices)
	merge.indices = indices
	sb := &strings.Builder{}
	for _, item := range indices {
		sb.WriteString(item)
	}
	suffix := ns
	if len(key) > 0 {
		suffix = ns + "." + key
	}
	merge.MergedIndex = m.cfg.IndexPrefix + "-" + metric + "-" + suffix + "-m-" + md5x.SumString(sb.String()).String16()
	merge.Indices = strings.Join(indices, ",")
}

// http://addon-elasticsearch.default.svc.cluster.local:9200/_tasks?detailed=true&actions=*reindex
func (m *IndexManager) doIndicesMerge(merge *MergeGroup, delete bool) (*elastic.BulkIndexByScrollResponse, error) {
	sources := elastic.NewReindexSource()
	sources.Index(merge.indices...)
	resp, err := m.client.Reindex().
		Source(sources).
		DestinationIndexAndType(merge.MergedIndex, m.cfg.IndexType).
		WaitForCompletion(true).
		Do(context.Background())
	if err != nil {
		err := fmt.Errorf("fail to reindex %v to %s : %s", merge.indices, merge.MergedIndex, err)
		m.log.Error(err)
		return nil, err
	}
	m.log.Infof("reindex %v to %s : %s", merge.indices, merge.MergedIndex, jsonx.MarshalAndIndent(resp))
	if delete {
		err = m.deleteIndices(merge.indices)
		if err != nil {
			return resp, fmt.Errorf("fail to remove indices: %v", merge.indices)
		}
	}
	return resp, nil
}
