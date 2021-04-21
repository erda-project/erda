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
