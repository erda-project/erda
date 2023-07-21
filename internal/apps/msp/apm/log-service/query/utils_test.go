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

package log_service

import (
	"testing"
)

func TestLogKeyGroup_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		g    LogKeyGroup
		want bool
	}{
		{
			name: "nil",
			g:    nil,
			want: true,
		},
		{
			name: "empty",
			g:    LogKeyGroup{},
			want: true,
		},
		{
			name: "not empty",
			g: LogKeyGroup{
				logServiceKey: StringList{"1"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
