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

package enhancedqueue

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/priorityqueue"
)

type SnapshotObj struct {
	Pending          json.RawMessage `json:"pending"`
	Processing       json.RawMessage `json:"processing"`
	ProcessingWindow int64           `json:"processingWindow"`
}

func (eq *EnhancedQueue) Export() json.RawMessage {
	eq.lock.Lock()
	defer eq.lock.Unlock()

	var obj SnapshotObj
	obj.Pending = eq.pending.Export()
	obj.Processing = eq.processing.Export()
	obj.ProcessingWindow = eq.processingWindow
	b, _ := json.Marshal(&obj)
	return b
}

func (eq *EnhancedQueue) Import(rawMsg json.RawMessage) error {
	eq.lock.Lock()
	defer eq.lock.Unlock()

	var obj SnapshotObj
	if err := json.Unmarshal(rawMsg, &obj); err != nil {
		return err
	}
	// restore pending
	pending := priorityqueue.NewPriorityQueue()
	if err := pending.Import(obj.Pending); err != nil {
		return fmt.Errorf("failed to import pending queue, err: %v", err)
	}
	eq.pending = pending
	// restore processing
	processing := priorityqueue.NewPriorityQueue()
	if err := processing.Import(obj.Processing); err != nil {
		return fmt.Errorf("failed to import processing queue, err: %v", err)
	}
	eq.processing = processing
	eq.processingWindow = obj.ProcessingWindow
	return nil
}
