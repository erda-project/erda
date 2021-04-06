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

package syncmap

import (
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
)

type StringInterfaceMap struct {
	sync.Map
}

func (m *StringInterfaceMap) MarshalJSON() ([]byte, error) {
	tmpMap := make(map[string]interface{})
	m.Range(func(k, v interface{}) bool {
		tmpMap[k.(string)] = MarkInterfaceType(v)
		return true
	})
	return json.Marshal(tmpMap)
}

func (m *StringInterfaceMap) UnmarshalJSON(data []byte) error {
	var tmpMap map[string]interface{}
	if err := json.Unmarshal(data, &tmpMap); err != nil {
		return err
	}
	for k, v := range tmpMap {
		m.Store(k, v)
	}
	return nil
}

func (m *StringInterfaceMap) Get(key string, o interface{}) error {
	v, ok := m.Load(key)
	if !ok {
		return errors.Errorf("not found")
	}

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, o); err != nil {
		return err
	}

	return nil
}

func (m *StringInterfaceMap) GetMap() map[string]interface{} {
	tmpMap := make(map[string]interface{})
	m.Range(func(k, v interface{}) bool {
		tmpMap[k.(string)] = v
		return true
	})
	return tmpMap
}

func (m *StringInterfaceMap) Store(k, v interface{}) {
	m.Map.Store(k, v)
}
