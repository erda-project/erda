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
	"container/heap"
	"sort"
)

// PriorityQueue 优先队列
type PriorityQueue struct {
	// 使用 data 再次封装，防止 Len() 等方法被暴露
	data *priorityQueue
}

// priorityQueue 优先队列
type priorityQueue struct {
	items     []Item
	itemByKey map[string]Item
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		data: &priorityQueue{
			items:     make([]Item, 0),
			itemByKey: make(map[string]Item),
		},
	}
}

// Get 根据 key 获取 BaseItem
func (pq *PriorityQueue) Get(key string) Item {
	return pq.data.itemByKey[key]
}

// Peek 获取优先级最高的 BaseItem
func (pq *PriorityQueue) Peek() Item {
	if len(pq.data.items) == 0 {
		return nil
	}
	return pq.data.items[0]
}

// Pop 弹出优先级最高的 BaseItem，并返回
func (pq *PriorityQueue) Pop() Item {
	if len(pq.data.items) == 0 {
		return nil
	}
	return heap.Pop(pq.data).(Item)
}

// Add 新增 BaseItem
func (pq *PriorityQueue) Add(item Item) {
	if res, ok := pq.data.itemByKey[item.Key()]; ok {
		if res.Priority() != item.Priority() {
			res.SetPriority(item.Priority())
			heap.Fix(pq.data, res.Index())
		}
	} else {
		heap.Push(pq.data, convertItem(item))
	}
	sort.Sort(pq.data)
}

// Remove 删除 BaseItem
func (pq *PriorityQueue) Remove(key string) Item {
	if _item, ok := pq.data.itemByKey[key]; ok {
		delete(pq.data.itemByKey, key)
		return heap.Remove(pq.data, _item.Index()).(Item)
	}
	return nil
}

// Len 返回队列长度
func (pq *PriorityQueue) Len() int {
	return pq.data.Len()
}

// Range range items and apply func to item one by one.
func (pq *PriorityQueue) Range(f func(Item) (stopRange bool)) {
	for _, item := range pq.data.items {
		stopRange := f(item)
		if stopRange {
			break
		}
	}
}

// LeftHasHigherOrder judge order of two items.
// return true if left has higher order.
func (pq *PriorityQueue) LeftHasHigherOrder(left, right string) bool {
	leftItem, leftExist := pq.data.itemByKey[left]
	rightItem, rightExist := pq.data.itemByKey[right]
	if !leftExist && !rightExist {
		return false
	}
	if !leftExist {
		return false
	}
	if !rightExist {
		return true
	}
	// both exist, judge by index
	return leftItem.Index() < rightItem.Index()
}
