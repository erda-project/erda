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
	data map[string]string
	mu   sync.RWMutex
}

func (m *Metadata) Add(key, value string) {
	if m.data == nil {
		m.init()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *Metadata) Get(key string) (string, bool) {
	if m.data == nil {
		m.init()
	}
	m.mu.RLock()
	defer m.mu.Unlock()
	v, ok := m.data[key]
	return v, ok
}

func (m *Metadata) Clone() *Metadata {
	m.mu.Lock()
	defer m.mu.Unlock()
	newdata := make(map[string]string, len(m.data))
	for k, v := range m.data {
		newdata[k] = v
	}
	return &Metadata{
		data: newdata,
	}
}

func (m *Metadata) init() {
	m.data = make(map[string]string)
}
