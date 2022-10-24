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

package util

import (
	"context"
	"testing"
)

func TestDisplayStatusText(t *testing.T) {
	type args struct {
		ctx    context.Context
		status string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty status",
			args: args{},
			want: "-",
		},
		{
			name: "non-exist status",
			args: args{
				status: "not-exist",
			},
			want: "-",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DisplayStatusText(tt.args.ctx, tt.args.status); got != tt.want {
				t.Errorf("DisplayStatusText() = %v, want %v", got, tt.want)
			}
		})
	}
}
