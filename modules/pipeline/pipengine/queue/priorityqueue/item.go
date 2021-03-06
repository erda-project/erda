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
