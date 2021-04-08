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

package units

import (
	"time"
)

// Converter .
type Converter interface {
	Valid(unit string) bool
	Convert(from, to string, value float64) (float64, bool)
}

var converts = []Converter{
	&TimeConverter{
		BaseConverter{
			Default: "ns",
			Units: map[string]float64{
				"ns":  float64(time.Nanosecond),
				"ms":  float64(time.Millisecond),
				"s":   float64(time.Second),
				"min": float64(time.Minute),
				"h":   float64(time.Hour),
				"d":   24 * float64(time.Hour),
			},
		},
	},
	&ByteConverter{
		BaseConverter{
			Default: "b",
			Units: map[string]float64{
				"b":  1,
				"B":  1,
				"kb": 1 * 1024,
				"KB": 1 * 1024,
				"mb": 1 * 1024 * 1024,
				"MB": 1 * 1024 * 1024,
				"gb": 1 * 1024 * 1024 * 1024,
				"GB": 1 * 1024 * 1024 * 1024,
				"tb": 1 * 1024 * 1024 * 1024 * 1024,
				"TB": 1 * 1024 * 1024 * 1024 * 1024,
			},
		},
	},
}

// BaseConverter .
type BaseConverter struct {
	Default string
	Units   map[string]float64
}

// Convert .
func (c *BaseConverter) Convert(from, to string, value float64) (float64, bool) {
	if from == "" {
		from = c.Default
	}
	base, ok := c.Units[from]
	if !ok {
		return value, false
	}
	val := value * base
	base, ok = c.Units[to]
	if !ok {
		return value, false
	}
	return val / base, true
}

// Valid .
func (c *BaseConverter) Valid(unit string) bool {
	_, ok := c.Units[unit]
	return ok
}

// TimeConverter .
type TimeConverter struct {
	BaseConverter
}

// ByteConverter .
type ByteConverter struct {
	BaseConverter
}

// Convert .
func Convert(from, to string, data interface{}) interface{} {
	val, ok := convertFloat64(data)
	if !ok {
		return data
	}
	for _, c := range converts {
		if c.Valid(to) {
			result, ok := c.Convert(from, to, val)
			if !ok {
				return data
			}
			return result
		}
	}
	return data
}

func convertFloat64(obj interface{}) (float64, bool) {
	switch val := obj.(type) {
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return float64(val), true
	}
	return 0, false
}
