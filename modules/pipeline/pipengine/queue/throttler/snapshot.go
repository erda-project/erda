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

package throttler

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/enhancedqueue"
)

type SnapshotObj struct {
	Name             string                     `json:"name"`
	QueueByName      map[string]json.RawMessage `json:"queueByName"`      // queueByName 无法从 keyRelatedQueues 中还原，因为可能存在空队列
	KeyRelatedQueues map[string][]string        `json:"keyRelatedQueues"` // 只关心加入了哪些队列，无需重复引用队列
}

func (t *throttler) Export() json.RawMessage {
	t.lock.Lock()
	defer t.lock.Unlock()

	obj := SnapshotObj{
		Name:             t.name,
		QueueByName:      make(map[string]json.RawMessage),
		KeyRelatedQueues: make(map[string][]string),
	}
	// queueByName
	for qName, queue := range t.queueByName {
		obj.QueueByName[qName] = queue.Export()
	}
	// keyRelatedQueues
	for key, queueByName := range t.keyRelatedQueues {
		for qName := range queueByName {
			obj.KeyRelatedQueues[key] = append(obj.KeyRelatedQueues[key], qName)
		}
	}
	b, _ := json.Marshal(&obj)
	return b
}

func (t *throttler) Import(rawMsg json.RawMessage) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	var obj SnapshotObj
	if err := json.Unmarshal(rawMsg, &obj); err != nil {
		return err
	}

	_nt := NewNamedThrottler(obj.Name, nil)
	nt := _nt.(*throttler)

	t.name = nt.name
	// queueByName
	t.queueByName = make(map[string]*enhancedqueue.EnhancedQueue)
	for qName, queueRawJSON := range obj.QueueByName {
		eq := enhancedqueue.NewEnhancedQueue(0)
		if err := eq.Import(queueRawJSON); err != nil {
			return fmt.Errorf("failed to import enhanced queue, queue name: %s, err: %v", qName, err)
		}
		t.queueByName[qName] = eq
	}
	// keyRelatedQueues
	t.keyRelatedQueues = make(map[string]map[string]*enhancedqueue.EnhancedQueue)
	for key, relatedQueues := range obj.KeyRelatedQueues {
		t.keyRelatedQueues[key] = make(map[string]*enhancedqueue.EnhancedQueue)
		for _, qName := range relatedQueues {
			t.keyRelatedQueues[key][qName] = t.queueByName[qName]
		}
	}

	return nil
}
