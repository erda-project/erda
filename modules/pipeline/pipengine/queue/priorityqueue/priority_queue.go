package priorityqueue

import (
	"container/heap"
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
