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

package logic

import (
	"context"
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func Test_getNetportalURL(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty ctx",
			args: args{
				ctx: context.Background(),
			},
			want: "",
		},
		{
			name: "ctx with netportal url",
			args: args{
				ctx: context.WithValue(context.Background(), apistructs.NETPORTAL_URL, "url1"),
			},
			want: "url1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNetportalURL(tt.args.ctx); got != tt.want {
				t.Errorf("getNetportalURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
