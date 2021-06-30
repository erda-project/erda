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

package endpoints

import "testing"

func Test_shouldCheckPermission(t *testing.T) {
	type args struct {
		isInternalClient       bool
		isInternalActionClient bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "isInternalClient",
			args: args{
				isInternalClient:       true,
				isInternalActionClient: false,
			},
			want: false,
		},
		{
			name: "isInternalActionClient",
			args: args{
				isInternalClient:       false,
				isInternalActionClient: true,
			},
			want: true,
		},
		{
			name: "isInternalClient_and_isInternalActionClient",
			args: args{
				isInternalClient:       true,
				isInternalActionClient: true,
			},
			want: true,
		},
		{
			name: "otherClient",
			args: args{
				isInternalClient:       false,
				isInternalActionClient: false,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldCheckPermission(tt.args.isInternalClient, tt.args.isInternalActionClient); got != tt.want {
				t.Errorf("shouldCheckPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}
