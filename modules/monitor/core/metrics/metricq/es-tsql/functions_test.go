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
)

var ctx Context

func Test_BuildInFunctions(t *testing.T) {
	tests := []struct {
		name    string
		args    []interface{}
		want    interface{}
		wantErr bool
	}{
		{"default_value", []interface{}{"1", "2", "3"}, "1", true},
		{"default_value", []interface{}{nil, "1"}, "1", false},
		{"default_value", []interface{}{"1", nil}, "1", false},
		{"format", []interface{}{}, "1", true},
		{"format", []interface{}{"1"}, "1", false},
		{"format", []interface{}{nil}, "", true},
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
