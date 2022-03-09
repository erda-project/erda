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

package common

import (
	"testing"
)

func TestIsJSONArray(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"", args{b: []byte(`[{"a":1}]`)}, true},
		{"", args{b: []byte(`{"a":1}`)}, false},
		{"", args{b: []byte(`[]`)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsJSONArray(tt.args.b); got != tt.want {
				t.Errorf("isJSONArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{key: "abc.edf/hij"},
			want: "abc_edf_hij",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeKey(tt.args.key); got != tt.want {
				t.Errorf("NormalizeKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
