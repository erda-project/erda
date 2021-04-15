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

package tsql

import (
	"testing"
	"time"
)

var ctx Context

func Test_BuildInFunctions(t *testing.T) {
	tests := []struct {
		name    string
		args    []interface{}
		want    interface{}
		wantErr bool
	}{
		{"default_value", []interface{}{"1", "2", "3"}, "error", true},
		{"default_value", []interface{}{nil, "1"}, "1", false},
		{"default_value", []interface{}{"1", nil}, "1", false},
		{"format", []interface{}{}, "error", true},
		{"format", []interface{}{"1"}, "1", false},
		{"format", []interface{}{nil}, "error", true},
		{"format_time", []interface{}{}, "error", true},
		{"format_time", []interface{}{"s", "s"}, "error", true},
		{"format_time", []interface{}{time.Now(), 7.6}, "error", true},
		{"format_time", []interface{}{"2000-01-01 15:04:05", "2006-01-02"}, "2000-01-01", false},
		{"format_date", []interface{}{}, "error", true},
		{"format_date", []interface{}{"string"}, "error", true},
		{"format_date", []interface{}{"2000-01-01 15:04:05"}, "2000-01-01", false},
		{"format_bytes", []interface{}{}, "error", true},
		{"format_bytes", []interface{}{1024}, "1KB", false},
		{"format_duration", []interface{}{}, "error", true},
		{"format_duration", []interface{}{1000100, "ms", 0}, "16m40s", false},
		{"format_duration", []interface{}{1000100, "s"}, "277h48m20s", false},
		{"format_duration", []interface{}{1000100, "m"}, "16668h20m0s", false},
		{"format_duration", []interface{}{1000100, "h"}, "1000100h0m0s", false},
		{"format_duration", []interface{}{101, "d"}, "2424h0m0s", false},
		{"format_duration", []interface{}{101, 1}, "error", true},
		{"map", []interface{}{}, "error", true},
		{"map", []interface{}{0, 0, "Gen0", 1}, "error", true},
		{"map", []interface{}{int64(1), int(0), "Gen0", int(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{int64(1), uint(0), "Gen0", uint(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{int64(1), uint64(0), "Gen0", uint64(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{int64(1), float32(0), "Gen0", float32(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{int64(1), float64(0), "Gen0", float64(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{int64(1), int64(0), "Gen0", int64(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{uint64(1), int(0), "Gen0", int(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{uint64(1), uint(0), "Gen0", uint(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{uint64(1), uint64(0), "Gen0", uint64(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{uint64(1), float32(0), "Gen0", float32(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{uint64(1), float64(0), "Gen0", float64(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{uint64(1), int64(0), "Gen0", int64(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{float64(1), int(0), "Gen0", int(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{float64(1), uint(0), "Gen0", uint(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{float64(1), uint64(0), "Gen0", uint64(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{float64(1), float32(0), "Gen0", float32(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{float64(1), float64(0), "Gen0", float64(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{float64(1), int64(0), "Gen0", int64(1), "Gen1"}, "Gen1", false},
		{"map", []interface{}{"1", int64(0), "Gen0", int64(1), "Gen1"}, "1", false},
		{"map", []interface{}{3, int64(0), "Gen0", int64(1), "Gen1"}, 3, false},
		{"round_float", []interface{}{}, "error", true},
		{"round_float", []interface{}{int64(7), 1}, int64(7), false},
		{"round_float", []interface{}{float32(7.85), 1}, 7.8, false},
		{"round_float", []interface{}{float64(7.85), 1}, 7.8, false},
		{"trim", []interface{}{"    trim    ", " "}, "trim", false},
		{"trim_left", []interface{}{"    trim", " "}, "trim", false},
		{"trim_right", []interface{}{"trim    ", " "}, "trim", false},
		{"trim_space", []interface{}{}, "error", true},
		{"trim_space", []interface{}{"        trim    "}, "trim", false},
		{"trim_space", []interface{}{1.1}, "trim", true},
		{"trim_prefix", []interface{}{"ttrim", "t"}, "trim", false},
		{"trim_suffix", []interface{}{"trimm", "m"}, "trim", false},
		{"max_value", []interface{}{}, "error", true},
		{"max_value", []interface{}{1, 2}, 2, false},
		{"max_value", []interface{}{3, 2}, 3, false},
		{"min_value", []interface{}{}, "error", true},
		{"min_value", []interface{}{1, 2}, 1, false},
		{"min_value", []interface{}{3, 2}, 2, false},
		{"min_value", []interface{}{}, "error", true},
		{"min_value", []interface{}{}, "error", true},
		{"parse_time", []interface{}{}, "error", true},
		{"parse_time", []interface{}{"2021-01-02 20:07:08", "2006-01"}, "error", true},
		{"substring", []interface{}{}, "error", true},
		{"substring", []interface{}{1, 2}, "error", true},
		{"substring", []interface{}{"substring", 3, 9}, "string", false},
		{"substring", []interface{}{"substring", 10, 10}, "", false},
		{"substring", []interface{}{"substring", 2, 1}, "", false},
		{"tostring", []interface{}{}, "error", true},
		{"tostring", []interface{}{nil}, "", false},
		{"tostring", []interface{}{1}, "1", false},
		{"if", []interface{}{}, "error", true},
		{"if", []interface{}{true, 1, 2}, 1, false},
		{"if", []interface{}{false, 1, 2}, 2, false},
		{"if", []interface{}{"if", 1, 2}, "error", true},
		{"eq", []interface{}{}, "error", true},
		{"eq", []interface{}{"eq", "eq"}, true, false},
		{"eq", []interface{}{"eq", "neq"}, false, false},
		{"neq", []interface{}{}, "error", true},
		{"neq", []interface{}{"eq", "eq"}, false, false},
		{"neq", []interface{}{"eq", "neq"}, true, false},
		{"include", []interface{}{}, "error", true},
		{"include", []interface{}{"include", "in", "clude"}, false, false},
		{"include", []interface{}{"include", "in", "clude", "include"}, true, false},
		{"gt", []interface{}{}, "error", true},
		{"gt", []interface{}{"t", 1}, "error", true},
		{"gt", []interface{}{2, 1}, true, false},
		{"gte", []interface{}{}, "error", true},
		{"gte", []interface{}{"t", 1}, "error", true},
		{"gte", []interface{}{1, 1}, true, false},
		{"lt", []interface{}{}, "error", true},
		{"lt", []interface{}{2, 1}, false, false},
		{"lt", []interface{}{"t", 1}, "error", true},
		{"lte", []interface{}{}, "error", true},
		{"lte", []interface{}{1, 1}, true, false},
		{"lte", []interface{}{"t", 1}, "error", true},
		{"andf", []interface{}{}, "error", true},
		{"andf", []interface{}{true, true}, true, false},
		{"andf", []interface{}{true, false, true}, false, false},
		{"andf", []interface{}{"t", true}, "error", true},
		{"orf", []interface{}{}, "error", true},
		{"orf", []interface{}{true, false}, true, false},
		{"orf", []interface{}{false, false}, false, false},
		{"orf", []interface{}{"t", true}, "error", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildInFunction := BuildInFunctions[tt.name]
			got, err := buildInFunction(ctx, tt.args...)
			if err != nil && tt.wantErr == true {
				return
			}
			if err != nil && tt.wantErr == false {
				t.Errorf("buildInFunction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildInFunction() = %v, want %v", got, tt.want)
			}
		})
	}
}
