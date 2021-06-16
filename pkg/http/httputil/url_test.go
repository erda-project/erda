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

package httputil

import "testing"

func TestJoinPath(t *testing.T) {
	type args struct {
		appendRoot bool
		segments   []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{},
			want: "",
		},
		{
			name: "appendRoot",
			args: args{
				appendRoot: true,
			},
			want: "/",
		},
		{
			name: "1",
			args: args{
				segments: []string{
					"path1",
				},
			},
			want: "path1",
		},
		{
			name: "1",
			args: args{
				appendRoot: true,
				segments: []string{
					"path1",
				},
			},
			want: "/path1",
		},
		{
			name: "3",
			args: args{
				appendRoot: true,
				segments: []string{
					"path1",
					"path2",
					"path3",
				},
			},
			want: "/path1/path2/path3",
		},
		{
			name: "escape",
			args: args{
				appendRoot: true,
				segments: []string{
					"path1",
					"/",
					"path2",
				},
			},
			want: "/path1/%2F/path2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JoinPath(tt.args.appendRoot, tt.args.segments...); got != tt.want {
				t.Errorf("JoinPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
