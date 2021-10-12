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

package manager

import "encoding/json"

type SnapshotObj struct {
	QueueUsageByID map[string]json.RawMessage `json:"queueUsageByID"`
}

func (mgr *defaultManager) Export() json.RawMessage {
	mgr.qLock.Lock()
	defer mgr.qLock.Unlock()
	obj := SnapshotObj{
		QueueUsageByID: make(map[string]json.RawMessage),
	}
	for qID, queue := range mgr.queueByID {
		u := queue.Usage()
		uByte, _ := json.Marshal(u)
		obj.QueueUsageByID[qID] = uByte
	}
	b, _ := json.Marshal(&obj)
	return b
}

// Import default queue manager execute in memory, don't need import
func (mgr *defaultManager) Import(rawMsg json.RawMessage) error {
	return nil
}
