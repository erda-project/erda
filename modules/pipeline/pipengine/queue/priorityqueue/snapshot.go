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

package priorityqueue

import (
	"encoding/json"
	"fmt"
	"time"
)

type SnapshotObj struct {
	Key          string    `json:"key"`
	Priority     int64     `json:"priority"`
	CreationTime time.Time `json:"creationTime"`
	Index        int       `json:"index"`
}

func (pq *PriorityQueue) Export() json.RawMessage {
	var objs []*SnapshotObj
	for _, item := range pq.data.items {
		obj := &SnapshotObj{
			Key:          item.Key(),
			Priority:     item.Priority(),
			CreationTime: item.CreationTime(),
			Index:        item.Index(),
		}
		objs = append(objs, obj)
	}
	b, _ := json.Marshal(&objs)
	return b
}

func (pq *PriorityQueue) Import(rawMsg json.RawMessage) error {
	var objs []*SnapshotObj
	if err := json.Unmarshal(rawMsg, &objs); err != nil {
		return fmt.Errorf("failed to import priority queue, err: %v", err)
	}
	// restore items && itemByKey
	pq.data.itemByKey = make(map[string]Item, len(objs))
	pq.data.items = make([]Item, len(objs))
	for i, obj := range objs {
		defaultItem := obj.convert2DefaultItem()
		pq.data.items[i] = defaultItem
		pq.data.itemByKey[obj.Key] = defaultItem
	}
	return nil
}

func (so *SnapshotObj) convert2DefaultItem() *defaultItem {
	return &defaultItem{
		key:          so.Key,
		priority:     so.Priority,
		creationTime: so.CreationTime,
		index:        so.Index,
	}
}
