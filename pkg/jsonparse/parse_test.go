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

package jsonparse

import (
	"math"
	"testing"
)

func TestJsonOneLine(t *testing.T) {
	tests := []struct {
		name string
		i    interface{}
		want string
	}{
		{name: "int", i: 1, want: "1"},
		{name: "int8 min", i: int8(math.MinInt8), want: "-128"},
		{name: "int8 max", i: int8(math.MaxInt8), want: "127"},
		{name: "int32 min", i: int32(math.MinInt32), want: "-2147483648"},
		{name: "int32 max", i: int32(math.MaxInt32), want: "2147483647"},
		{name: "int64 min", i: int64(math.MinInt64), want: "-9223372036854775808"},
		{name: "int64 max", i: int64(math.MaxInt64), want: "9223372036854775807"},
		{name: "uint", i: uint(1), want: "1"},
		{name: "uint8 max", i: uint8(math.MaxUint8), want: "255"},
		{name: "uint32 max", i: uint32(math.MaxUint32), want: "4294967295"},
		{name: "uint64 max", i: uint64(math.MaxUint64), want: "18446744073709551615"},
		{name: "float64", i: float64(132455555555555.1), want: "132455555555555.1"},
		{name: "[]byte", i: []byte{'a', 'b'}, want: "ab"},
		{name: "string", i: "ab", want: "ab"},
		{name: "json", i: "[{\"aaa\":111}]", want: "[{\"aaa\":111}]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JsonOneLine(tt.i); got != tt.want {
				t.Errorf("JsonOneLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
