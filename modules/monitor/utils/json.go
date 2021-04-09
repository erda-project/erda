// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

type JSONMap map[string]interface{}

// Scanner 接口实现方法
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(map[string]interface{})
		return nil
	}

	var bytes []byte
	convertBytes(value, &bytes)
	if err := json.Unmarshal(bytes, &m); err != nil {
		return err
	}
	return nil
}

// driver Valuer 接口实现方法
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		m = make(map[string]interface{})
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

type NullMap struct {
	Map   map[string]interface{}
	Valid bool
}

// Scanner 接口实现方法
func (nm *NullMap) Scan(value interface{}) error {
	if value == nil {
		nm.Map, nm.Valid = nil, false
		return nil
	}

	var bytes []byte
	convertBytes(value, &bytes)
	if err := json.Unmarshal(bytes, &nm.Map); err != nil {
		return err
	}
	nm.Valid = true
	return nil
}

// driver Valuer 接口实现方法
func (nm NullMap) Value() (driver.Value, error) {
	if !nm.Valid {
		return nil, nil
	}
	return nm.Map, nil
}

// 转换为bytes
func convertBytes(src interface{}, dest *[]byte) error {
	if dest == nil {
		return errors.New("convert bytes err: dest is nil")
	}

	switch s := src.(type) {
	case string:
		*dest = []byte(s)
	case []byte:
		*dest = cloneBytes(s)
	case nil:
		*dest = nil
	default:
		*dest = nil
	}
	return nil
}

// 拷贝bytes
func cloneBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func IsJsonArray(b []byte) bool {
	x := bytes.TrimLeft(b, " \t\r\n")
	return len(x) > 0 && x[0] == '['
}
