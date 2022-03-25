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

package odata

import (
	"sync"
)

type Metadata struct {
	Data map[string]string `json:"Data"`
	mu   sync.RWMutex
}

func NewMetadata() *Metadata {
	return &Metadata{Data: map[string]string{}}
}

func (m *Metadata) Add(key, value string) {
	if m.Data == nil {
		m.init()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data[key] = value
}

func (m *Metadata) Get(key string) (string, bool) {
	if m.Data == nil {
		m.init()
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.Data[key]
	return v, ok
}

func (m *Metadata) Clone() *Metadata {
	m.mu.Lock()
	defer m.mu.Unlock()
	newdata := make(map[string]string, len(m.Data))
	for k, v := range m.Data {
		newdata[k] = v
	}
	return &Metadata{
		Data: newdata,
	}
}

func (m *Metadata) init() {
	m.Data = make(map[string]string)
}
