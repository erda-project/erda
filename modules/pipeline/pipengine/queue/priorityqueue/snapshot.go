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
