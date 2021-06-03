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

package uc

import "testing"

func Test_getDomain(t *testing.T) {
	type args struct {
		host    string
		domains []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				host: "erda-org.app.terminus.io",
				domains: []string{
					".terminus.io",
					".erda.cloud",
				},
			},
			want: ".terminus.io",
		},
		{
			args: args{
				host: "erda-org.app.erda.cloud",
				domains: []string{
					".terminus.io",
					".erda.cloud",
				},
			},
			want: ".erda.cloud",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDomain(tt.args.host, tt.args.domains); got != tt.want {
				t.Errorf("getDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}
