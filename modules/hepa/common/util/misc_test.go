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

package util

import (
	"reflect"
	"testing"
)

func TestUniqStringSlice(t *testing.T) {
	type args struct {
		slice []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"normal",
			args{[]string{"e", "a", "c", "c", "a", "d", "b", "d", "b"}},
			[]string{"a", "b", "c", "d", "e"},
		},
		{
			"edge1",
			args{[]string{}},
			[]string{},
		},
		{
			"edge2",
			args{[]string{"a"}},
			[]string{"a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UniqStringSlice(tt.args.slice); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
