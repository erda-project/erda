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

package openai_v1_models

import "testing"

func TestParseModelUUIDFromDisplayName(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				s: "test",
			},
			want: "",
		},
		{
			name: "test2",
			args: args{
				s: "test [ID:123]",
			},
			want: "123",
		},
		{
			name: "test3",
			args: args{
				s: "test [ID:123][ID:456]",
			},
			want: "123",
		},
		{
			name: "test4",
			args: args{
				s: "[id:abcd]",
			},
			want: "abcd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseModelUUIDFromDisplayName(tt.args.s); got != tt.want {
				t.Errorf("ParseModelUUIDFromDisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}
