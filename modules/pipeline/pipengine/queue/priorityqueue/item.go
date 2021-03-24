package priorityqueue

import "time"

// Item 是优先队列存储的结构
type Item interface {
	Key() string
	Priority() int64
	SetPriority(int64)
	CreationTime() time.Time
	Index() int
	SetIndex(int)
}

func NewItem(key string, priority int64, creationTime time.Time) Item {
	item := &defaultItem{
		key:          key,
		priority:     priority,
		creationTime: creationTime,
	}
	return item
}

type defaultItem struct {
	key          string
	priority     int64
	creationTime time.Time
	index        int
}

func (d *defaultItem) Key() string                { return d.key }
func (d *defaultItem) Priority() int64            { return d.priority }
func (d *defaultItem) SetPriority(priority int64) { d.priority = priority }
func (d *defaultItem) CreationTime() time.Time    { return d.creationTime }
func (d *defaultItem) Index() int                 { return d.index }
func (d *defaultItem) SetIndex(index int)         { d.index = index }

func convertItem(item Item) *defaultItem {
	return &defaultItem{
		key:          item.Key(),
		priority:     item.Priority(),
		creationTime: item.CreationTime().Round(0),
	}
}
